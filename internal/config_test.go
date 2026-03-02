// internal/config_test.go
package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkerConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "workers", "dev-001", "config.yaml")

	original := &WorkerConfig{
		WorkerID:     "dev-001",
		Role:         "dev",
		Provider:     "claude",
		DefaultModel: "claude-sonnet-4-6",
		PaneID:       "42",
		CreatedAt:    "2026-02-24T10:00:00Z",
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
	if loaded.Provider != original.Provider {
		t.Errorf("Provider = %q, want %q", loaded.Provider, original.Provider)
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

func TestWorkerConfigRoleScope(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	original := &WorkerConfig{
		WorkerID:  "dev-001",
		Role:      "dev",
		RoleScope: "global",
		RolePath:  "/home/user/.agents/roles/dev",
		Provider:  "claude",
		PaneID:    "42",
		CreatedAt: "2026-03-02T10:00:00Z",
	}

	if err := original.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadWorkerConfig(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.RoleScope != "global" {
		t.Errorf("RoleScope = %q, want global", loaded.RoleScope)
	}
	if loaded.RolePath != "/home/user/.agents/roles/dev" {
		t.Errorf("RolePath = %q, want /home/user/.agents/roles/dev", loaded.RolePath)
	}
}

func TestWorkerConfigRoleScopeOmitEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &WorkerConfig{
		WorkerID: "dev-001",
		Role:     "dev",
		Provider: "claude",
		PaneID:   "42",
	}

	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "role_scope") {
		t.Error("YAML should not contain role_scope when empty")
	}
	if strings.Contains(content, "role_path") {
		t.Error("YAML should not contain role_path when empty")
	}
}
