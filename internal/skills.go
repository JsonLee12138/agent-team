// internal/skills.go
package internal

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// roleYAML is the minimal struct to extract skills from role.yaml.
type roleYAML struct {
	Skills []string `yaml:"skills"`
}

// ReadRoleSkillsFromPath reads the skills list from a role's references/role.yaml at the given path.
func ReadRoleSkillsFromPath(rolePath string) ([]string, error) {
	yamlPath := filepath.Join(rolePath, "references", "role.yaml")
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read role.yaml: %w", err)
	}
	var ry roleYAML
	if err := yaml.Unmarshal(data, &ry); err != nil {
		return nil, fmt.Errorf("parse role.yaml: %w", err)
	}
	return ry.Skills, nil
}

// ReadRoleSkills reads the skills list from a role's references/role.yaml.
func ReadRoleSkills(root, roleName string) ([]string, error) {
	return ReadRoleSkillsFromPath(RoleDir(root, roleName))
}

// CopySkillsToWorktreeFromPath copies the role skill and its dependency skills
// into the worktree's .claude/skills/ and .codex/skills/ directories, using rolePath directly.
func CopySkillsToWorktreeFromPath(wtPath, root, roleName, rolePath string) error {
	// Collect skill directories to copy
	type skillSource struct {
		name string
		path string
	}
	var sources []skillSource

	// 1. Copy the role skill itself from rolePath
	if _, err := os.Stat(rolePath); err == nil {
		sources = append(sources, skillSource{name: roleName, path: rolePath})
	}

	// 2. Read dependency skills from role.yaml
	skills, err := ReadRoleSkillsFromPath(rolePath)
	if err != nil {
		return err
	}

	for _, skillName := range skills {
		skillPath := findSkillPath(root, skillName)
		if skillPath != "" {
			destName := parseSkillName(skillName)
			sources = append(sources, skillSource{name: destName, path: skillPath})
		} else {
			fmt.Fprintf(os.Stderr, "Warning: skill '%s' not found in any search path, skipping\n", skillName)
		}
	}

	// Copy to all provider skill directories
	targets := []string{
		filepath.Join(wtPath, ".claude", "skills"),
		filepath.Join(wtPath, ".codex", "skills"),
		filepath.Join(wtPath, ".opencode", "skills"),
		filepath.Join(wtPath, ".gemini", "skills"),
	}

	for _, targetBase := range targets {
		for _, src := range sources {
			dst := filepath.Join(targetBase, src.name)
			if err := copyDir(src.path, dst); err != nil {
				return fmt.Errorf("copy skill %s to %s: %w", src.name, dst, err)
			}
		}
	}

	return nil
}

// CopySkillsToWorktree copies the role skill and its dependency skills
// into the worktree's .claude/skills/ and .codex/skills/ directories.
func CopySkillsToWorktree(wtPath, root, roleName string) error {
	return CopySkillsToWorktreeFromPath(wtPath, root, roleName, RoleDir(root, roleName))
}

// parseSkillName extracts the short skill name from formats like
// "antfu/skills@vite" → "vite", or returns the original name if no "@" is present.
func parseSkillName(skillName string) string {
	if idx := strings.LastIndex(skillName, "@"); idx >= 0 {
		return skillName[idx+1:]
	}
	return skillName
}

// findSkillPath searches for a skill in known locations.
// Supports both plain names ("vite") and scoped names ("antfu/skills@vite").
func findSkillPath(root, skillName string) string {
	// Build candidate names: full name first, then short name (after @) if different
	candidates := []string{skillName}
	shortName := parseSkillName(skillName)
	if shortName != skillName {
		candidates = append(candidates, shortName)
	}

	searchDirs := buildSearchDirs(root)

	for _, name := range candidates {
		for _, dir := range searchDirs {
			p := filepath.Join(dir, name)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	// 远程下载回退（仅 scoped 格式）
	if shortName != skillName {
		targetDir := filepath.Join(root, ".claude", "skills")
		if downloaded := tryRemoteDownload(skillName, targetDir, shortName); downloaded != "" {
			return downloaded
		}
	}

	return ""
}

// pluginSkillsDir 返回 Plugin 内置技能目录（从环境变量读取）
func pluginSkillsDir() string {
	if root := os.Getenv("CLAUDE_PLUGIN_ROOT"); root != "" {
		return filepath.Join(root, "skills")
	}
	return ""
}

// buildSearchDirs 构建 5 层技能搜索目录
func buildSearchDirs(root string) []string {
	var dirs []string
	if d := pluginSkillsDir(); d != "" {
		dirs = append(dirs, d) // 层 1: Plugin 内置
	}
	dirs = append(dirs, filepath.Join(ResolveAgentsDir(root), "teams")) // 层 2: .agents/teams
	dirs = append(dirs, filepath.Join(root, "skills"))                  // 层 3: project/skills
	dirs = append(dirs, filepath.Join(root, ".claude", "skills"))       // 层 4: .claude/skills
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".claude", "skills")) // 层 5: ~/.claude/skills
	}
	return dirs
}

// tryRemoteDownload 尝试通过 npx skills install 下载远程技能（非阻塞，失败只 warn）
func tryRemoteDownload(skillName, targetDir, shortName string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "npx", "skills", "install", skillName, "--target", targetDir)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: remote skill download failed '%s': %v\n", skillName, err)
		return ""
	}
	downloaded := filepath.Join(targetDir, shortName)
	if _, err := os.Stat(downloaded); err == nil {
		return downloaded
	}
	return ""
}

// providerToAgent maps provider names to npx skills add -a parameter values.
var providerToAgent = map[string]string{
	"claude":   "claude-code",
	"codex":    "codex",
	"opencode": "opencode",
	"gemini":   "gemini",
}

// skillTargetDir returns the skill installation target directory for a provider.
func skillTargetDir(wtPath, provider string) string {
	switch provider {
	case "claude":
		return filepath.Join(wtPath, ".claude", "skills")
	case "codex":
		return filepath.Join(wtPath, ".codex", "skills")
	case "opencode":
		return filepath.Join(wtPath, ".opencode", "skills")
	case "gemini":
		return filepath.Join(wtPath, ".gemini", "skills")
	default:
		return filepath.Join(wtPath, ".claude", "skills")
	}
}

// isScopedSkill returns true if the skill name contains "/" (scoped format like "antfu/skills@vite").
func isScopedSkill(skillName string) bool {
	return strings.Contains(skillName, "/")
}

// runNpxSkillsAdd runs "npx skills add <source> -a <agent> -y" in the given directory.
func runNpxSkillsAdd(cwd, skillName, provider string) error {
	agent, ok := providerToAgent[provider]
	if !ok {
		agent = "claude-code"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "npx", "skills", "add", skillName, "-a", agent, "-y")
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// InstallSkillsForWorkerFromPath installs role skills into the worktree using an explicit rolePath.
func InstallSkillsForWorkerFromPath(wtPath, root, roleName, rolePath, provider string) error {
	targetDir := skillTargetDir(wtPath, provider)

	// 1. Copy the role skill itself (always local)
	if _, err := os.Stat(rolePath); err == nil {
		dst := filepath.Join(targetDir, roleName)
		if err := copyDir(rolePath, dst); err != nil {
			return fmt.Errorf("copy role skill %s: %w", roleName, err)
		}
	}

	// 2. Read dependency skills from role.yaml
	skills, err := ReadRoleSkillsFromPath(rolePath)
	if err != nil {
		return err
	}

	for _, skillName := range skills {
		shortName := parseSkillName(skillName)
		dst := filepath.Join(targetDir, shortName)

		// Skip if already installed in target directory
		if _, err := os.Stat(dst); err == nil {
			fmt.Printf("  Skill '%s' already installed, skipping\n", shortName)
			continue
		}

		if isScopedSkill(skillName) {
			// Scoped: directly use npx skills add
			fmt.Printf("  Installing skill '%s' via npx...\n", skillName)
			if err := runNpxSkillsAdd(wtPath, skillName, provider); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: npx skills add '%s' failed: %v\n", skillName, err)
			}
		} else {
			// Plain: try local search first
			skillPath := findSkillPath(root, skillName)
			if skillPath != "" {
				if err := copyDir(skillPath, dst); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: copy skill '%s' failed: %v\n", skillName, err)
				}
			} else {
				// Fallback to npx skills add
				fmt.Printf("  Skill '%s' not found locally, trying npx...\n", skillName)
				if err := runNpxSkillsAdd(wtPath, skillName, provider); err != nil {
					fmt.Fprintf(os.Stderr, "Error: skill '%s' not found locally and npx install failed: %v\n", skillName, err)
				}
			}
		}
	}

	return nil
}

// InstallSkillsForWorker installs role skills into the worktree.
// - scoped names (containing "/") → npx skills add <source> -a <agent> -y
// - plain names → local 5-layer search copy → fallback npx skills add → error
// The role skill itself is always copied locally.
func InstallSkillsForWorker(wtPath, root, roleName, provider string) error {
	return InstallSkillsForWorkerFromPath(wtPath, root, roleName, RoleDir(root, roleName), provider)
}

// FindSkillPathPublic searches for a skill in known locations (exported wrapper).
func FindSkillPathPublic(root, skillName string) string {
	return findSkillPath(root, skillName)
}
func CopyDirPublic(src, dst string) error {
	return copyDir(src, dst)
}

// BridgeSkillsForProvider creates symlinks from .agents/teams/ and ~/.agents/roles/
// into the provider's skill scan directory (e.g., .claude/skills/, .opencode/skills/).
// Only directories containing SKILL.md are bridged. Existing entries are not overwritten.
func BridgeSkillsForProvider(cwd, provider string) error {
	targetDir := skillTargetDir(cwd, provider)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create target dir: %w", err)
	}

	// Source directories: project-level + global
	sources := []string{
		filepath.Join(ResolveAgentsDir(cwd), "teams"),
	}
	if globalDir, err := GlobalRolesDir(); err == nil {
		sources = append(sources, globalDir)
	}

	bridged := 0
	for _, srcDir := range sources {
		entries, err := os.ReadDir(srcDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			src := filepath.Join(srcDir, entry.Name())
			// Only bridge directories that contain SKILL.md
			if _, err := os.Stat(filepath.Join(src, "SKILL.md")); os.IsNotExist(err) {
				continue
			}

			dest := filepath.Join(targetDir, entry.Name())
			// Skip if already exists (don't overwrite manual installs or prior copies)
			if _, err := os.Lstat(dest); err == nil {
				continue
			}

			if err := os.Symlink(src, dest); err != nil {
				fmt.Fprintf(os.Stderr, "[agent-team] bridge-skills: symlink %s → %s: %v\n", src, dest, err)
				continue
			}
			bridged++
		}
	}

	if bridged > 0 {
		fmt.Fprintf(os.Stderr, "[agent-team] bridge-skills: linked %d skill(s) to %s\n", bridged, targetDir)
	}
	return nil
}

// copyDir recursively copies a directory.
func copyDir(src, dst string) error {
	// Remove destination if it exists to ensure a clean copy
	os.RemoveAll(dst)

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}
