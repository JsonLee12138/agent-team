// internal/config.go
package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// WorkerConfig represents an employee instance of a role.
type WorkerConfig struct {
	WorkerID         string `yaml:"worker_id"`
	Role             string `yaml:"role"`
	RoleScope        string `yaml:"role_scope,omitempty"`      // "project" | "global"
	RolePath         string `yaml:"role_path,omitempty"`       // absolute path for global roles
	Provider         string `yaml:"provider"`
	DefaultModel     string `yaml:"default_model,omitempty"`
	MainSessionID    string `yaml:"main_session_id,omitempty"`
	PaneID           string `yaml:"pane_id"`
	ControllerPaneID string `yaml:"controller_pane_id,omitempty"`
	CreatedAt        string `yaml:"created_at"`
	WorktreeCreated  *bool  `yaml:"worktree_created,omitempty"`
}

// MainSessionConfig stores the project's main/controller pane metadata.
type MainSessionConfig struct {
	Backend   string `yaml:"backend,omitempty"`
	PaneID    string `yaml:"pane_id"`
	UpdatedAt string `yaml:"updated_at,omitempty"`
}

// WorkerYAMLPath returns the path to worker.yaml in the worktree root.
func WorkerYAMLPath(wtPath string) string {
	return filepath.Join(wtPath, "worker.yaml")
}

// WorkerConfigDir returns the path to the centralized worker config directory.
func WorkerConfigDir(root, workerID string) string {
	return filepath.Join(ResolveAgentsDir(root), "workers", workerID)
}

// WorkerConfigPath returns the path to the centralized worker config file.
func WorkerConfigPath(root, workerID string) string {
	return filepath.Join(WorkerConfigDir(root, workerID), "worker.yaml")
}

// MainSessionYAMLPath returns the path to the project-local main pane config.
func MainSessionYAMLPath(root string) string {
	return filepath.Join(root, ".agent-team", "main-session.yaml")
}

func LoadWorkerConfig(path string) (*WorkerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read worker config %s: %w", path, err)
	}
	var cfg WorkerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse worker config %s: %w", path, err)
	}
	return &cfg, nil
}

func (c *WorkerConfig) IsWorktreeCreated() bool {
	if c == nil || c.WorktreeCreated == nil {
		return true
	}
	return *c.WorktreeCreated
}

// LoadWorkerConfigByID loads worker config from the new centralized path first,
// then falls back to the legacy worktree-local path.
func LoadWorkerConfigByID(root, wtBase, workerID string) (*WorkerConfig, string, error) {
	newPath := WorkerConfigPath(root, workerID)
	if cfg, err := LoadWorkerConfig(newPath); err == nil {
		return cfg, newPath, nil
	}

	legacyPath := WorkerYAMLPath(WtPath(root, wtBase, workerID))
	cfg, err := LoadWorkerConfig(legacyPath)
	if err != nil {
		return nil, "", fmt.Errorf("read centralized config %s or legacy config %s: %w", newPath, legacyPath, err)
	}
	return cfg, legacyPath, nil
}

func LoadMainSessionConfig(path string) (*MainSessionConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read main session config %s: %w", path, err)
	}
	var cfg MainSessionConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse main session config %s: %w", path, err)
	}
	return &cfg, nil
}

func (c *WorkerConfig) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal worker config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (c *MainSessionConfig) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal main session config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
