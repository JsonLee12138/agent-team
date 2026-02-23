// internal/config.go
package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type RoleConfig struct {
	Name            string `yaml:"name"`
	Description     string `yaml:"description"`
	DefaultProvider string `yaml:"default_provider"`
	DefaultModel    string `yaml:"default_model"`
	CreatedAt       string `yaml:"created_at"`
	PaneID          string `yaml:"pane_id"`
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
