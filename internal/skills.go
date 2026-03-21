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
	"sync"
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

// Deprecated: CopySkillsToWorktreeFromPath copies skills via full directory copy.
// Use InstallSkillsForWorkerFromPath with symlink mode instead.
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

// parseSkillName extracts the short skill name from scoped formats:
//   - "antfu/skills@vite" → "vite"       (org/repo@skill)
//   - "better-auth/better-icons" → "better-icons"  (org/skill)
//   - "vite" → "vite"                     (plain name)
func parseSkillName(skillName string) string {
	if idx := strings.LastIndex(skillName, "@"); idx >= 0 {
		return skillName[idx+1:]
	}
	if idx := strings.LastIndex(skillName, "/"); idx >= 0 {
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
		targetDir := filepath.Join(ResolveAgentsDir(root), ".cache", "skills")
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
	dirs = append(dirs, filepath.Join(ResolveAgentsDir(root), "teams")) // 层 2: .agent-team/teams
	dirs = append(dirs, filepath.Join(root, "skills"))                  // 层 3: project/skills
	dirs = append(dirs, filepath.Join(ResolveAgentsDir(root), ".cache", "skills")) // 层 4: project cache
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
	return filepath.Join(wtPath, skillTargetSuffix(provider))
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

// backgroundCheckOnce ensures only one background skill check runs per process lifetime.
var backgroundCheckOnce sync.Once

// backgroundSkillCheck silently runs "npx skills check" in the project root.
// Outputs update hints to stderr without blocking the caller.
// Uses sync.Once to avoid duplicate checks from concurrent worker creation.
func backgroundSkillCheck(root, provider string) {
	backgroundCheckOnce.Do(func() {
		agent, ok := providerToAgent[provider]
		if !ok {
			agent = "claude-code"
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "npx", "skills", "check", "-a", agent)
		cmd.Dir = root

		out, err := cmd.CombinedOutput()
		if err != nil {
			// Silently ignore errors — this is a best-effort check
			return
		}
		output := strings.TrimSpace(string(out))
		if output != "" {
			fmt.Fprintf(os.Stderr, "\n[skill update check]\n%s\n", output)
		}
	})
}

// InstallSkillsForWorkerFromPath installs role skills into the worktree using symlinks
// to project-level cache. When fresh is true, cached skills are re-installed.
func InstallSkillsForWorkerFromPath(wtPath, root, roleName, rolePath, provider string, fresh bool) error {
	// 1. Symlink the role skill itself (always local)
	if _, err := os.Stat(rolePath); err == nil {
		if err := symlinkSkill(wtPath, provider, roleName, rolePath); err != nil {
			return fmt.Errorf("link role skill %s: %w", roleName, err)
		}
	}

	// 2. Read dependency skills from role.yaml
	skills, err := ReadRoleSkillsFromPath(rolePath)
	if err != nil {
		return err
	}

	var cachedSkills []string

	for _, skillName := range skills {
		shortName := parseSkillName(skillName)
		dst := filepath.Join(skillTargetDir(wtPath, provider), shortName)

		// Skip if already installed in worktree (and not a broken symlink)
		if info, err := os.Stat(dst); err == nil && info != nil {
			fmt.Printf("  Skill '%s' already installed, skipping\n", shortName)
			continue
		}

		// Check project-level cache
		cachePath := projectSkillPath(root, provider, shortName)
		cacheHit := false
		if !fresh {
			if _, err := os.Stat(cachePath); err == nil {
				cacheHit = true
			}
		}

		if cacheHit {
			cachedSkills = append(cachedSkills, skillName)
			// Cache hit: symlink worktree → project cache
			if err := symlinkSkill(wtPath, provider, shortName, cachePath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: symlink cached skill '%s' failed: %v\n", shortName, err)
			}
		} else if isScopedSkill(skillName) {
			// Scoped: npx skills add to project root, then move to cache, then symlink
			fmt.Printf("  Installing skill '%s' via npx (project cache)...\n", skillName)
			if err := runNpxSkillsAdd(root, skillName, provider); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: npx skills add '%s' failed: %v\n", skillName, err)
				continue
			}
			// Move from npx output (<root>/.<provider>/skills/) to project cache
			if err := moveNpxResultToCache(root, provider, shortName); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: move skill '%s' to cache failed: %v\n", shortName, err)
				continue
			}
			if _, err := os.Stat(cachePath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: npx installed '%s' but expected path not found: %s\n", skillName, cachePath)
				continue
			}
			if err := symlinkSkill(wtPath, provider, shortName, cachePath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: symlink skill '%s' failed: %v\n", shortName, err)
			}
		} else {
			// Plain: try local search first
			skillPath := findSkillPath(root, skillName)
			if skillPath != "" {
				// Found locally: symlink directly to source
				if err := symlinkSkill(wtPath, provider, shortName, skillPath); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: symlink skill '%s' failed: %v\n", shortName, err)
				}
			} else {
				// Fallback: npx skills add to project root, then move to cache, then symlink
				fmt.Printf("  Skill '%s' not found locally, trying npx (project cache)...\n", skillName)
				if err := runNpxSkillsAdd(root, skillName, provider); err != nil {
					fmt.Fprintf(os.Stderr, "Error: skill '%s' not found locally and npx install failed: %v\n", skillName, err)
					continue
				}
				// Move from npx output to project cache
				if err := moveNpxResultToCache(root, provider, shortName); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: move skill '%s' to cache failed: %v\n", shortName, err)
					continue
				}
				if _, err := os.Stat(cachePath); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: npx installed '%s' but expected path not found: %s\n", skillName, cachePath)
					continue
				}
				if err := symlinkSkill(wtPath, provider, shortName, cachePath); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: symlink skill '%s' failed: %v\n", shortName, err)
				}
			}
		}
	}

	// Background update check: if any skills were served from cache, silently check for updates
	if len(cachedSkills) > 0 && !fresh {
		go backgroundSkillCheck(root, provider)
	}

	return nil
}

// InstallSkillsForWorker installs role skills into the worktree.
func InstallSkillsForWorker(wtPath, root, roleName, provider string) error {
	return InstallSkillsForWorkerFromPath(wtPath, root, roleName, RoleDir(root, roleName), provider, false)
}

// FindSkillPathPublic searches for a skill in known locations (exported wrapper).
func FindSkillPathPublic(root, skillName string) string {
	return findSkillPath(root, skillName)
}
func CopyDirPublic(src, dst string) error {
	return copyDir(src, dst)
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

// skillTargetSuffix returns the provider-specific skill directory suffix (e.g. ".claude/skills").
func skillTargetSuffix(provider string) string {
	switch provider {
	case "codex":
		return filepath.Join(".codex", "skills")
	case "opencode":
		return filepath.Join(".opencode", "skills")
	case "gemini":
		return filepath.Join(".gemini", "skills")
	default:
		return filepath.Join(".claude", "skills")
	}
}

// projectSkillPath returns the project-level skill cache path for a given skill.
// e.g. <root>/.agent-team/.cache/skills/<skillName>
// The provider parameter is kept for call-site compatibility but ignored internally.
func projectSkillPath(root, provider, skillName string) string {
	return filepath.Join(ResolveAgentsDir(root), ".cache", "skills", skillName)
}

// symlinkSkill creates a symlink in the worktree pointing to targetPath.
// Falls back to copyDir if symlink creation fails (e.g. Windows without Developer Mode).
func symlinkSkill(wtPath, provider, skillName, targetPath string) error {
	linkPath := filepath.Join(skillTargetDir(wtPath, provider), skillName)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(linkPath), 0755); err != nil {
		return err
	}

	// Remove existing entry (symlink, dir, or file)
	if err := os.RemoveAll(linkPath); err != nil {
		return fmt.Errorf("remove existing skill '%s': %w", skillName, err)
	}

	// Use absolute path for symlink target
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		absTarget = targetPath
	}

	if err := os.Symlink(absTarget, linkPath); err != nil {
		// Fallback to copy on platforms that don't support symlinks
		fmt.Fprintf(os.Stderr, "Warning: symlink not supported, falling back to copy for '%s'\n", skillName)
		return copyDir(targetPath, linkPath)
	}
	return nil
}

// moveNpxResultToCache moves the skill installed by npx from <root>/.agent-team/skills/<name>
// to the project cache at <root>/.agent-team/.cache/skills/<name>.
// npx skills add installs real files to <cwd>/.agent-team/skills/ and creates symlinks
// in <cwd>/.<provider>/skills/. We relocate the real files and clean up the symlinks.
func moveNpxResultToCache(root, provider, shortName string) error {
	// npx installs real files to .agent-team/skills/<name>
	npxPath := filepath.Join(ResolveAgentsDir(root), "skills", shortName)
	cachePath := projectSkillPath(root, provider, shortName)

	if _, err := os.Stat(npxPath); err != nil {
		return fmt.Errorf("npx result not found at %s: %w", npxPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return err
	}
	os.RemoveAll(cachePath)

	if err := os.Rename(npxPath, cachePath); err != nil {
		// Cross-device fallback: copy then remove
		if cpErr := copyDir(npxPath, cachePath); cpErr != nil {
			return cpErr
		}
		os.RemoveAll(npxPath)
	}

	// Clean up the symlink npx created in .<provider>/skills/
	npxLink := filepath.Join(root, skillTargetSuffix(provider), shortName)
	if isSymlink(npxLink) {
		os.Remove(npxLink)
	}

	return nil
}

// isSymlink checks if the given path is a symbolic link.
func isSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// CachedSkillUsage describes which workers are using a cached skill.
type CachedSkillUsage struct {
	SkillName string   // cache directory name
	Workers   []string // worker IDs with symlinks pointing to this cache entry
}

// FindCachedSkillUsage scans all worktrees to find which cached skills are actively
// symlinked from worker skill directories. Returns a map of skill name → worker IDs.
func FindCachedSkillUsage(root, wtBase string) map[string][]string {
	cacheDir := filepath.Join(ResolveAgentsDir(root), ".cache", "skills")
	absCacheDir, err := filepath.Abs(cacheDir)
	if err != nil {
		return nil
	}

	// Collect all cached skill names
	cacheEntries, err := os.ReadDir(cacheDir)
	if err != nil {
		return nil
	}
	cachedNames := make(map[string]bool, len(cacheEntries))
	for _, e := range cacheEntries {
		cachedNames[e.Name()] = true
	}
	if len(cachedNames) == 0 {
		return nil
	}

	usage := make(map[string][]string)

	// Scan all worktrees
	wtDir := filepath.Join(root, wtBase)
	workers, err := os.ReadDir(wtDir)
	if err != nil {
		return usage
	}

	providers := []string{"claude", "codex", "opencode", "gemini"}
	for _, w := range workers {
		if !w.IsDir() {
			continue
		}
		workerID := w.Name()
		wtPath := filepath.Join(wtDir, workerID)

		for _, prov := range providers {
			skillDir := skillTargetDir(wtPath, prov)
			skills, err := os.ReadDir(skillDir)
			if err != nil {
				continue
			}
			for _, s := range skills {
				linkPath := filepath.Join(skillDir, s.Name())
				if !isSymlink(linkPath) {
					continue
				}
				target, err := os.Readlink(linkPath)
				if err != nil {
					continue
				}
				// Check if symlink target points into the cache directory
				if strings.HasPrefix(target, absCacheDir+string(os.PathSeparator)) {
					skillName := s.Name()
					if cachedNames[skillName] {
						// Deduplicate worker IDs per skill
						found := false
						for _, id := range usage[skillName] {
							if id == workerID {
								found = true
								break
							}
						}
						if !found {
							usage[skillName] = append(usage[skillName], workerID)
						}
					}
				}
			}
		}
	}

	return usage
}
