// internal/config_test.go
package internal

import (
	"path/filepath"
	"testing"
)

func TestWorkerConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "workers", "dev-001", "config.yaml")

	original := &WorkerConfig{
		WorkerID:        "dev-001",
		Role:            "dev",
		DefaultProvider: "claude",
		DefaultModel:    "claude-sonnet-4-6",
		PaneID:          "42",
		CreatedAt:       "2026-02-24T10:00:00Z",
	}

	if err := original.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadWorkerConfig(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.WorkerID != original.WorkerID {
		t.Errorf("WorkerID = %q, want %q", loaded.WorkerID, original.WorkerID)
	}
	if loaded.Role != original.Role {
		t.Errorf("Role = %q, want %q", loaded.Role, original.Role)
	}
	if loaded.DefaultProvider != original.DefaultProvider {
		t.Errorf("DefaultProvider = %q, want %q", loaded.DefaultProvider, original.DefaultProvider)
	}
	if loaded.PaneID != original.PaneID {
		t.Errorf("PaneID = %q, want %q", loaded.PaneID, original.PaneID)
	}
}

func TestLoadWorkerConfigNotFound(t *testing.T) {
	_, err := LoadWorkerConfig("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestWorkerConfigSaveUpdate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &WorkerConfig{WorkerID: "test-001", Role: "test", PaneID: ""}
	cfg.Save(path)

	cfg.PaneID = "99"
	cfg.Save(path)

	reloaded, _ := LoadWorkerConfig(path)
	if reloaded.PaneID != "99" {
		t.Errorf("PaneID after update = %q, want %q", reloaded.PaneID, "99")
	}
}
