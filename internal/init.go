package internal

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// DetectedProvider holds info about a detected agent provider binary.
type DetectedProvider struct {
	Name string
	Path string
}

// DetectInstalledProviders checks for known agent provider binaries on PATH.
func DetectInstalledProviders() []DetectedProvider {
	names := []string{"claude", "gemini", "opencode", "codex"}
	var detected []DetectedProvider
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			detected = append(detected, DetectedProvider{Name: name, Path: path})
		}
	}
	return detected
}

// PluginRoleCandidate represents a role found in the plugin's skills directory.
type PluginRoleCandidate struct {
	Name    string
	Path    string
	DirHash string
}

// ScanPluginRoles scans $CLAUDE_PLUGIN_ROOT/skills/ for directories containing references/role.yaml.
func ScanPluginRoles() []PluginRoleCandidate {
	pluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	if pluginRoot == "" {
		return nil
	}
	skillsDir := filepath.Join(pluginRoot, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil
	}

	var candidates []PluginRoleCandidate
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		rolePath := filepath.Join(skillsDir, e.Name())
		yamlPath := filepath.Join(rolePath, "references", "role.yaml")
		if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
			continue
		}
		hash, _ := HashLocalRoleDir(rolePath)
		candidates = append(candidates, PluginRoleCandidate{
			Name:    e.Name(),
			Path:    rolePath,
			DirHash: hash,
		})
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Name < candidates[j].Name
	})
	return candidates
}

// HashLocalRoleDir computes a SHA-256 hash of all files in a role directory.
// Files are sorted by relative path for deterministic output.
func HashLocalRoleDir(dir string) (string, error) {
	h := sha256.New()
	var paths []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		paths = append(paths, rel)
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Strings(paths)
	for _, rel := range paths {
		data, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%s\n%d\n", rel, len(data))
		h.Write(data)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// InstallAction describes what happened during a plugin role install.
type InstallAction string

const (
	InstallActionSkipped   InstallAction = "skipped"
	InstallActionInstalled InstallAction = "installed"
	InstallActionUpdated   InstallAction = "updated"
)

// InstallPluginRoleToGlobal copies a plugin role to the global roles directory.
// It compares hashes to decide: skip (identical), install (new), or update (changed).
func InstallPluginRoleToGlobal(candidate PluginRoleCandidate, globalDir string) (InstallAction, error) {
	destDir := filepath.Join(globalDir, candidate.Name)

	// Check if destination already exists
	if _, err := os.Stat(destDir); err == nil {
		// Destination exists — compare hashes
		existingHash, err := HashLocalRoleDir(destDir)
		if err != nil {
			return "", fmt.Errorf("hash existing role %s: %w", candidate.Name, err)
		}
		if existingHash == candidate.DirHash {
			return InstallActionSkipped, nil
		}
		// Hash differs — update (remove old, copy new)
		if err := os.RemoveAll(destDir); err != nil {
			return "", fmt.Errorf("remove old role %s: %w", candidate.Name, err)
		}
		if err := copyDir(candidate.Path, destDir); err != nil {
			return "", fmt.Errorf("update role %s: %w", candidate.Name, err)
		}
		return InstallActionUpdated, nil
	}

	// New installation
	if err := copyDir(candidate.Path, destDir); err != nil {
		return "", fmt.Errorf("install role %s: %w", candidate.Name, err)
	}
	return InstallActionInstalled, nil
}

// InitProject creates the .agents/teams/ directory with a .gitkeep file.
func InitProject(root string) error {
	teamsDir := filepath.Join(root, ".agents", "teams")
	if err := os.MkdirAll(teamsDir, 0755); err != nil {
		return fmt.Errorf("create .agents/teams/: %w", err)
	}
	gitkeep := filepath.Join(teamsDir, ".gitkeep")
	if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
		if err := os.WriteFile(gitkeep, []byte(""), 0644); err != nil {
			return fmt.Errorf("create .gitkeep: %w", err)
		}
	}
	return nil
}

// EnsureGlobalRolesDir creates ~/.agents/roles/ if it doesn't exist.
func EnsureGlobalRolesDir() (string, error) {
	dir, err := GlobalRolesDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create global roles dir: %w", err)
	}
	return dir, nil
}

// FormatProviderList returns a comma-separated string of provider names.
func FormatProviderList(providers []DetectedProvider) string {
	names := make([]string, len(providers))
	for i, p := range providers {
		names[i] = p.Name
	}
	return strings.Join(names, ", ")
}
