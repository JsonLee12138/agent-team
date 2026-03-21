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
	WorkerID         string     `yaml:"worker_id"`
	Role             string     `yaml:"role"`
	RoleScope        string     `yaml:"role_scope,omitempty"`      // "project" | "global"
	RolePath         string     `yaml:"role_path,omitempty"`       // absolute path for global roles
	Provider         string     `yaml:"provider"`
	DefaultModel     string     `yaml:"default_model,omitempty"`
	MainSessionID    string     `yaml:"main_session_id,omitempty"`
	PaneID           string     `yaml:"pane_id"`
	ControllerPaneID string     `yaml:"controller_pane_id,omitempty"`
	TaskID           string     `yaml:"task_id,omitempty"`
	TaskPath         string     `yaml:"task_path,omitempty"`
	Status           TaskStatus `yaml:"status,omitempty"`
	CreatedAt        string     `yaml:"created_at"`
	UpdatedAt        string     `yaml:"updated_at,omitempty"`
	WorktreeCreated  *bool      `yaml:"worktree_created,omitempty"`
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

// LoadWorkerConfigByID loads worker config from the worktree-local path first,
// then falls back to the centralized compatibility path.
func LoadWorkerConfigByID(root, wtBase, workerID string) (*WorkerConfig, string, error) {
	localPath := WorkerYAMLPath(WtPath(root, wtBase, workerID))
	if cfg, err := LoadWorkerConfig(localPath); err == nil {
		return cfg, localPath, nil
	}

	compatPath := WorkerConfigPath(root, workerID)
	cfg, err := LoadWorkerConfig(compatPath)
	if err != nil {
		return nil, "", fmt.Errorf("read local config %s or centralized config %s: %w", localPath, compatPath, err)
	}
	return cfg, compatPath, nil
}

func WorkerConfigWritePath(root, wtBase, workerID string) string {
	wtPath := WtPath(root, wtBase, workerID)
	if info, err := os.Stat(wtPath); err == nil && info.IsDir() {
		return WorkerYAMLPath(wtPath)
	}
	return WorkerConfigPath(root, workerID)
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
