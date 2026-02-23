// internal/config_test.go
package internal

import (
	"path/filepath"
	"testing"
)

func TestRoleConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	original := &RoleConfig{
		Name:            "backend",
		Description:     "Backend developer",
		DefaultProvider: "claude",
		DefaultModel:    "claude-sonnet-4-6",
		CreatedAt:       "2026-02-24T10:00:00Z",
		PaneID:          "42",
	}

	if err := original.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadRoleConfig(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Name != original.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, original.Name)
	}
	if loaded.DefaultProvider != original.DefaultProvider {
		t.Errorf("DefaultProvider = %q, want %q", loaded.DefaultProvider, original.DefaultProvider)
	}
	if loaded.PaneID != original.PaneID {
		t.Errorf("PaneID = %q, want %q", loaded.PaneID, original.PaneID)
	}
}

func TestLoadRoleConfigNotFound(t *testing.T) {
	_, err := LoadRoleConfig("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestRoleConfigSaveUpdate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &RoleConfig{Name: "test", PaneID: ""}
	cfg.Save(path)

	cfg.PaneID = "99"
	cfg.Save(path)

	reloaded, _ := LoadRoleConfig(path)
	if reloaded.PaneID != "99" {
		t.Errorf("PaneID after update = %q, want %q", reloaded.PaneID, "99")
	}
}
