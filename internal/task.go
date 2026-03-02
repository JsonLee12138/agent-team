// internal/task.go
package internal

import (
	"time"

	"gopkg.in/yaml.v3"
)

type ChangeStatus string

const (
	ChangeStatusDraft        ChangeStatus = "draft"
	ChangeStatusAssigned     ChangeStatus = "assigned"
	ChangeStatusImplementing ChangeStatus = "implementing"
	ChangeStatusVerifying    ChangeStatus = "verifying"
	ChangeStatusDone         ChangeStatus = "done"
	ChangeStatusArchived     ChangeStatus = "archived"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusSkipped    TaskStatus = "skipped"
)

// Task represents a unit of work within a Change.
type Task struct {
	ID     int        `yaml:"id"`
	Title  string     `yaml:"title"`
	Status TaskStatus `yaml:"status"`
}

// VerifyConfig defines how to verify a Change.
type VerifyConfig struct {
	Command string `yaml:"command,omitempty"`
	Timeout string `yaml:"timeout,omitempty"` // "300s", "2m"
	Skip    bool   `yaml:"skip,omitempty"`
}

// ParseTimeout parses the timeout string into a time.Duration.
func (v VerifyConfig) ParseTimeout() (time.Duration, error) {
	if v.Timeout == "" {
		return 5 * time.Minute, nil // default
	}
	return time.ParseDuration(v.Timeout)
}

// Change represents a named change with tasks and verification config.
type Change struct {
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	Status      ChangeStatus `yaml:"status"`
	CreatedAt   string       `yaml:"created_at"`
	AssignedTo  string       `yaml:"assigned_to,omitempty"`
	Tasks       []Task       `yaml:"tasks,omitempty"`
	Verify      VerifyConfig `yaml:"verify,omitempty"`
}

// TaskConfig is the top-level configuration for tasks directory.
type TaskConfig struct {
	Version  int                `yaml:"version"`
	Defaults TaskConfigDefaults `yaml:"defaults"`
}

// TaskConfigDefaults contains default settings for all changes.
type TaskConfigDefaults struct {
	Verify    VerifyConfig `yaml:"verify"`
	Lifecycle []string     `yaml:"lifecycle"`
}

// DefaultTaskConfig returns the default TaskConfig.
func DefaultTaskConfig() TaskConfig {
	return TaskConfig{
		Version: 1,
		Defaults: TaskConfigDefaults{
			Verify: VerifyConfig{
				Timeout: "5m",
			},
			Lifecycle: []string{"draft", "assigned", "implementing", "verifying", "done", "archived"},
		},
	}
}

// MarshalYAML marshals Change to YAML.
func (c *Change) MarshalYAML() (interface{}, error) {
	type changeAlias Change
	return (*changeAlias)(c), nil
}

// UnmarshalYAML unmarshals Change from YAML.
func (c *Change) UnmarshalYAML(value *yaml.Node) error {
	type changeAlias Change
	alias := (*changeAlias)(c)
	return value.Decode(alias)
}

// ValidChangeStatus returns true if the status is valid.
func ValidChangeStatus(s ChangeStatus) bool {
	switch s {
	case ChangeStatusDraft, ChangeStatusAssigned, ChangeStatusImplementing,
		ChangeStatusVerifying, ChangeStatusDone, ChangeStatusArchived:
		return true
	default:
		return false
	}
}

// ValidTaskStatus returns true if the status is valid.
func ValidTaskStatus(s TaskStatus) bool {
	switch s {
	case TaskStatusPending, TaskStatusInProgress, TaskStatusDone, TaskStatusSkipped:
		return true
	default:
		return false
	}
}
