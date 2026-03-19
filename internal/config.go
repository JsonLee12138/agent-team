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
	RoleScope        string `yaml:"role_scope,omitempty"` // "project" | "global"
	RolePath         string `yaml:"role_path,omitempty"`  // absolute path for global roles
	Provider         string `yaml:"provider"`
	DefaultModel     string `yaml:"default_model,omitempty"`
	MainSessionID    string `yaml:"main_session_id,omitempty"`
	PaneID           string `yaml:"pane_id"`
	ControllerPaneID string `yaml:"controller_pane_id,omitempty"`
	CreatedAt        string `yaml:"created_at"`
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
