package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultNormalizeInterval         = 2 * time.Minute
	DefaultVisibilityRefreshInterval = 10 * time.Second
)

// CatalogPipelineConfig controls normalize and visibility polling settings.
type CatalogPipelineConfig struct {
	Version    int                     `yaml:"version"`
	Normalize  CatalogNormalizeConfig  `yaml:"normalize"`
	Visibility CatalogVisibilityConfig `yaml:"visibility"`
}

// CatalogNormalizeConfig configures the normalize polling loop.
type CatalogNormalizeConfig struct {
	Interval string `yaml:"interval"`
}

// CatalogVisibilityConfig configures the visibility refresh loop.
type CatalogVisibilityConfig struct {
	RefreshInterval string `yaml:"refresh_interval"`
}

// DefaultCatalogPipelineConfig returns the default pipeline config.
func DefaultCatalogPipelineConfig() CatalogPipelineConfig {
	return CatalogPipelineConfig{
		Version: 1,
		Normalize: CatalogNormalizeConfig{
			Interval: DefaultNormalizeInterval.String(),
		},
		Visibility: CatalogVisibilityConfig{
			RefreshInterval: DefaultVisibilityRefreshInterval.String(),
		},
	}
}

// CatalogPipelineConfigPath returns the config path under .agents/.
func CatalogPipelineConfigPath(root string) string {
	return filepath.Join(root, ".agents", "catalog-pipeline.yaml")
}

// LoadCatalogPipelineConfig loads config or returns defaults if missing.
func LoadCatalogPipelineConfig(root string) (*CatalogPipelineConfig, error) {
	path := CatalogPipelineConfigPath(root)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultCatalogPipelineConfig()
			return &cfg, nil
		}
		return nil, fmt.Errorf("read catalog pipeline config: %w", err)
	}
	var cfg CatalogPipelineConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse catalog pipeline config: %w", err)
	}
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	applyCatalogPipelineDefaults(&cfg)
	return &cfg, nil
}

// EnsureCatalogPipelineConfig writes defaults if missing and returns config.
func EnsureCatalogPipelineConfig(root string) (*CatalogPipelineConfig, error) {
	path := CatalogPipelineConfigPath(root)
	if _, err := os.Stat(path); err == nil {
		return LoadCatalogPipelineConfig(root)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("stat catalog pipeline config: %w", err)
	}

	cfg := DefaultCatalogPipelineConfig()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create catalog config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal catalog pipeline config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return nil, fmt.Errorf("write catalog pipeline config: %w", err)
	}
	return &cfg, nil
}

// NormalizeIntervalDuration parses the normalize interval.
func (c CatalogPipelineConfig) NormalizeIntervalDuration() (time.Duration, error) {
	if c.Normalize.Interval == "" {
		return DefaultNormalizeInterval, nil
	}
	d, err := time.ParseDuration(c.Normalize.Interval)
	if err != nil {
		return 0, fmt.Errorf("parse normalize interval: %w", err)
	}
	return d, nil
}

// VisibilityRefreshIntervalDuration parses the visibility refresh interval.
func (c CatalogPipelineConfig) VisibilityRefreshIntervalDuration() (time.Duration, error) {
	if c.Visibility.RefreshInterval == "" {
		return DefaultVisibilityRefreshInterval, nil
	}
	d, err := time.ParseDuration(c.Visibility.RefreshInterval)
	if err != nil {
		return 0, fmt.Errorf("parse visibility refresh interval: %w", err)
	}
	return d, nil
}

func applyCatalogPipelineDefaults(cfg *CatalogPipelineConfig) {
	if cfg.Normalize.Interval == "" {
		cfg.Normalize.Interval = DefaultNormalizeInterval.String()
	}
	if cfg.Visibility.RefreshInterval == "" {
		cfg.Visibility.RefreshInterval = DefaultVisibilityRefreshInterval.String()
	}
}
