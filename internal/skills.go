// internal/skills.go
package internal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// roleYAML is the minimal struct to extract skills from role.yaml.
type roleYAML struct {
	Skills []string `yaml:"skills"`
}

// ReadRoleSkills reads the skills list from a role's references/role.yaml.
func ReadRoleSkills(root, roleName string) ([]string, error) {
	yamlPath := RoleYAMLPath(root, roleName)
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

// CopySkillsToWorktree copies the role skill and its dependency skills
// into the worktree's .claude/skills/ and .codex/skills/ directories.
func CopySkillsToWorktree(wtPath, root, roleName string) error {
	// Collect skill directories to copy
	type skillSource struct {
		name string
		path string
	}
	var sources []skillSource

	// 1. Copy the role skill itself from agents/teams/<role>/
	roleDir := RoleDir(root, roleName)
	if _, err := os.Stat(roleDir); err == nil {
		sources = append(sources, skillSource{name: roleName, path: roleDir})
	}

	// 2. Read dependency skills from role.yaml
	skills, err := ReadRoleSkills(root, roleName)
	if err != nil {
		return err
	}

	for _, skillName := range skills {
		skillPath := findSkillPath(root, skillName)
		if skillPath != "" {
			// Use the short name as the target directory name
			// e.g., "antfu/skills@vite" → "vite"
			destName := parseSkillName(skillName)
			sources = append(sources, skillSource{name: destName, path: skillPath})
		} else {
			fmt.Fprintf(os.Stderr, "Warning: skill '%s' not found in any search path, skipping\n", skillName)
		}
	}

	// Copy to both .claude/skills/ and .codex/skills/
	targets := []string{
		filepath.Join(wtPath, ".claude", "skills"),
		filepath.Join(wtPath, ".codex", "skills"),
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

	for _, name := range candidates {
		// Project-local agents/teams/
		teamDir := filepath.Join(root, "agents", "teams", name)
		if _, err := os.Stat(teamDir); err == nil {
			return teamDir
		}

		// Project-local skills/
		local := filepath.Join(root, "skills", name)
		if _, err := os.Stat(local); err == nil {
			return local
		}

		// Global ~/.claude/skills/
		home, err := os.UserHomeDir()
		if err == nil {
			global := filepath.Join(home, ".claude", "skills", name)
			if _, err := os.Stat(global); err == nil {
				return global
			}
		}
	}

	return ""
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
