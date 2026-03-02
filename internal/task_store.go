// internal/task_store.go
package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// TasksDir returns the path to the .tasks directory.
func TasksDir(wtPath string) string {
	return filepath.Join(wtPath, ".tasks")
}

// TasksChangesDir returns the path to the .tasks/changes directory.
func TasksChangesDir(wtPath string) string {
	return filepath.Join(TasksDir(wtPath), "changes")
}

// TasksConfigPath returns the path to .tasks/config.yaml.
func TasksConfigPath(wtPath string) string {
	return filepath.Join(TasksDir(wtPath), "config.yaml")
}

// ChangeDirPath returns the path to a change directory.
func ChangeDirPath(wtPath, changeName string) string {
	return filepath.Join(TasksChangesDir(wtPath), changeName)
}

// ChangeYAMLPath returns the path to a change's change.yaml.
func ChangeYAMLPath(wtPath, changeName string) string {
	return filepath.Join(ChangeDirPath(wtPath, changeName), "change.yaml")
}

// InitTasksDir initializes the .tasks directory structure with a default config.
// This is idempotent — it will not overwrite an existing config.
func InitTasksDir(wtPath string) error {
	changesDir := TasksChangesDir(wtPath)

	// Create directories
	if err := os.MkdirAll(changesDir, 0755); err != nil {
		return fmt.Errorf("create tasks directories: %w", err)
	}

	// Initialize config.yaml if it doesn't exist
	configPath := TasksConfigPath(wtPath)
	if _, err := os.Stat(configPath); err == nil {
		return nil // config already exists
	}

	cfg := DefaultTaskConfig()
	if err := SaveTaskConfig(wtPath, &cfg); err != nil {
		return fmt.Errorf("save default config: %w", err)
	}

	return nil
}

// LoadTaskConfig loads the config from .tasks/config.yaml.
func LoadTaskConfig(wtPath string) (*TaskConfig, error) {
	configPath := TasksConfigPath(wtPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg TaskConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

// SaveTaskConfig saves the config to .tasks/config.yaml.
func SaveTaskConfig(wtPath string, cfg *TaskConfig) error {
	configPath := TasksConfigPath(wtPath)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// CreateTaskChange creates a new change directory with change.yaml.
// Optionally writes proposal.md and design.md if provided.
func CreateTaskChange(wtPath, changeName, description, proposal, design string) (string, error) {
	changeDir := ChangeDirPath(wtPath, changeName)

	// Create change directory
	if err := os.MkdirAll(changeDir, 0755); err != nil {
		return "", fmt.Errorf("create change directory: %w", err)
	}

	// Create change.yaml
	change := &Change{
		Name:        changeName,
		Description: description,
		Status:      ChangeStatusDraft,
		CreatedAt:   time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	}

	if err := SaveChange(wtPath, change); err != nil {
		return "", err
	}

	// Write proposal.md if provided
	if proposal != "" {
		proposalPath := filepath.Join(changeDir, "proposal.md")
		if err := os.WriteFile(proposalPath, []byte(proposal), 0644); err != nil {
			return "", fmt.Errorf("write proposal.md: %w", err)
		}
	}

	// Write design.md if provided
	if design != "" {
		designPath := filepath.Join(changeDir, "design.md")
		if err := os.WriteFile(designPath, []byte(design), 0644); err != nil {
			return "", fmt.Errorf("write design.md: %w", err)
		}
	}

	return changeDir, nil
}

// LoadChange loads a change from its change.yaml.
func LoadChange(wtPath, changeName string) (*Change, error) {
	yamlPath := ChangeYAMLPath(wtPath, changeName)
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("read change.yaml: %w", err)
	}

	var change Change
	if err := yaml.Unmarshal(data, &change); err != nil {
		return nil, fmt.Errorf("parse change.yaml: %w", err)
	}

	return &change, nil
}

// SaveChange saves a change to its change.yaml.
func SaveChange(wtPath string, change *Change) error {
	yamlPath := ChangeYAMLPath(wtPath, change.Name)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(yamlPath), 0755); err != nil {
		return fmt.Errorf("create change directory: %w", err)
	}

	data, err := yaml.Marshal(change)
	if err != nil {
		return fmt.Errorf("marshal change: %w", err)
	}

	if err := os.WriteFile(yamlPath, data, 0644); err != nil {
		return fmt.Errorf("write change.yaml: %w", err)
	}

	return nil
}

// ListChanges returns all changes, sorted by name.
func ListChanges(wtPath string) ([]*Change, error) {
	changesDir := TasksChangesDir(wtPath)

	// Check if .tasks/changes exists
	if _, err := os.Stat(changesDir); os.IsNotExist(err) {
		return []*Change{}, nil
	}

	entries, err := os.ReadDir(changesDir)
	if err != nil {
		return nil, fmt.Errorf("read changes directory: %w", err)
	}

	var changes []*Change
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		change, err := LoadChange(wtPath, e.Name())
		if err != nil {
			// Skip broken changes
			continue
		}

		changes = append(changes, change)
	}

	// Sort by name (time.Time prefix will naturally sort chronologically)
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Name < changes[j].Name
	})

	return changes, nil
}

// ListActiveChanges returns all changes that are not archived.
func ListActiveChanges(wtPath string) ([]*Change, error) {
	allChanges, err := ListChanges(wtPath)
	if err != nil {
		return nil, err
	}

	var active []*Change
	for _, change := range allChanges {
		if change.Status != ChangeStatusArchived {
			active = append(active, change)
		}
	}

	return active, nil
}

// CountActiveChanges returns the number of non-archived changes.
func CountActiveChanges(wtPath string) int {
	active, err := ListActiveChanges(wtPath)
	if err != nil {
		return 0
	}
	return len(active)
}
