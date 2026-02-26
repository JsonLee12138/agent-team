// internal/config.go
package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type RoleConfig struct {
	Name             string `yaml:"name"`
	Description      string `yaml:"description"`
	DefaultProvider  string `yaml:"default_provider"`
	DefaultModel     string `yaml:"default_model"`
	CreatedAt        string `yaml:"created_at"`
	PaneID           string `yaml:"pane_id"`
	ControllerPaneID string `yaml:"controller_pane_id,omitempty"`
}

func LoadRoleConfig(path string) (*RoleConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg RoleConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return &cfg, nil
}

func (c *RoleConfig) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// WorkerConfig represents an employee instance of a role.
type WorkerConfig struct {
	WorkerID         string `yaml:"worker_id"`
	Role             string `yaml:"role"`
	DefaultProvider  string `yaml:"default_provider"`
	DefaultModel     string `yaml:"default_model"`
	PaneID           string `yaml:"pane_id"`
	ControllerPaneID string `yaml:"controller_pane_id,omitempty"`
	CreatedAt        string `yaml:"created_at"`
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
