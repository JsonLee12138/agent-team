// internal/task_test.go
package internal

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestVerifyConfigParseTimeout(t *testing.T) {
	tests := []struct {
		name    string
		cfg     VerifyConfig
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "empty timeout defaults to 5m",
			cfg:     VerifyConfig{Timeout: ""},
			want:    5 * time.Minute,
			wantErr: false,
		},
		{
			name:    "300s parses correctly",
			cfg:     VerifyConfig{Timeout: "300s"},
			want:    300 * time.Second,
			wantErr: false,
		},
		{
			name:    "2m parses correctly",
			cfg:     VerifyConfig{Timeout: "2m"},
			want:    2 * time.Minute,
			wantErr: false,
		},
		{
			name:    "invalid format error",
			cfg:     VerifyConfig{Timeout: "invalid"},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cfg.ParseTimeout()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimeout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeYAMLSerialization(t *testing.T) {
	change := &Change{
		Name:        "2024-01-01-test-feature",
		Description: "Test Feature",
		Status:      ChangeStatusDraft,
		CreatedAt:   "2024-01-01T12:00:00Z",
		AssignedTo:  "worker1",
		Tasks: []Task{
			{ID: 1, Title: "Task 1", Status: TaskStatusPending},
			{ID: 2, Title: "Task 2", Status: TaskStatusInProgress},
		},
		Verify: VerifyConfig{
			Command: "go test ./...",
			Timeout: "5m",
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(change)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal back
	var restored Change
	if err := yaml.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify round-trip
	if restored.Name != change.Name {
		t.Errorf("Name mismatch: %s != %s", restored.Name, change.Name)
	}
	if restored.Status != change.Status {
		t.Errorf("Status mismatch: %s != %s", restored.Status, change.Status)
	}
	if len(restored.Tasks) != len(change.Tasks) {
		t.Errorf("Tasks length mismatch: %d != %d", len(restored.Tasks), len(change.Tasks))
	}
	if restored.Verify.Command != change.Verify.Command {
		t.Errorf("Verify.Command mismatch: %s != %s", restored.Verify.Command, change.Verify.Command)
	}
}

func TestDefaultTaskConfig(t *testing.T) {
	cfg := DefaultTaskConfig()
	if cfg.Version != 1 {
		t.Errorf("Version mismatch: %d != 1", cfg.Version)
	}
	if cfg.Defaults.Verify.Timeout != "5m" {
		t.Errorf("Default timeout mismatch: %s != 5m", cfg.Defaults.Verify.Timeout)
	}
	if len(cfg.Defaults.Lifecycle) == 0 {
		t.Error("Lifecycle should not be empty")
	}
}

func TestValidChangeStatus(t *testing.T) {
	tests := []struct {
		status ChangeStatus
		valid  bool
	}{
		{ChangeStatusDraft, true},
		{ChangeStatusAssigned, true},
		{ChangeStatusImplementing, true},
		{ChangeStatusVerifying, true},
		{ChangeStatusDone, true},
		{ChangeStatusArchived, true},
		{ChangeStatus("invalid"), false},
	}

	for _, tt := range tests {
		if ValidChangeStatus(tt.status) != tt.valid {
			t.Errorf("ValidChangeStatus(%s) = %v, want %v", tt.status, ValidChangeStatus(tt.status), tt.valid)
		}
	}
}

func TestValidTaskStatus(t *testing.T) {
	tests := []struct {
		status TaskStatus
		valid  bool
	}{
		{TaskStatusPending, true},
		{TaskStatusInProgress, true},
		{TaskStatusDone, true},
		{TaskStatusSkipped, true},
		{TaskStatus("invalid"), false},
	}

	for _, tt := range tests {
		if ValidTaskStatus(tt.status) != tt.valid {
			t.Errorf("ValidTaskStatus(%s) = %v, want %v", tt.status, ValidTaskStatus(tt.status), tt.valid)
		}
	}
}
