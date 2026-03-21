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
	path := filepath.Join(dir, ".worktrees", "dev-001", "worker.yaml")

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

func TestWorkerConfigPathHelpers(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".worktrees", "dev-001"), 0755); err != nil {
		t.Fatalf("mkdir .agents: %v", err)
	}
	if got := WorkerConfigDir(dir, "dev-001"); got != filepath.Join(dir, ".worktrees", "dev-001") {
		t.Fatalf("WorkerConfigDir = %q", got)
	}
	if got := WorkerConfigPath(dir, "dev-001"); got != filepath.Join(dir, ".worktrees", "dev-001", "worker.yaml") {
		t.Fatalf("WorkerConfigPath = %q", got)
	}
}

func TestWorkerConfigIsWorktreeCreatedCompatibility(t *testing.T) {
	trueValue := true
	falseValue := false
	if !((&WorkerConfig{}).IsWorktreeCreated()) {
		t.Fatal("nil worktree_created should default to true")
	}
	if !((&WorkerConfig{WorktreeCreated: &trueValue}).IsWorktreeCreated()) {
		t.Fatal("true worktree_created should stay true")
	}
	if (&WorkerConfig{WorktreeCreated: &falseValue}).IsWorktreeCreated() {
		t.Fatal("false worktree_created should stay false")
	}
}

func TestLoadWorkerConfigByIDPrefersLocalPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".worktrees", "dev-001"), 0755); err != nil {
		t.Fatalf("mkdir .agents: %v", err)
	}
	localCfg := &WorkerConfig{WorkerID: "dev-001", Role: "local-role"}
		localPath := WorkerYAMLPath(filepath.Join(dir, ".worktrees", "dev-001"))
	if err := localCfg.Save(localPath); err != nil {
		t.Fatalf("save local config: %v", err)
	}
	loaded, path, err := LoadWorkerConfigByID(dir, ".worktrees", "dev-001")
	if err != nil {
		t.Fatalf("LoadWorkerConfigByID: %v", err)
	}
	if loaded.Role != "local-role" {
		t.Fatalf("Role = %q, want local-role", loaded.Role)
	}
	if path != localPath {
		t.Fatalf("path = %q", path)
	}
}

func TestLoadWorkerConfigByIDUsesLocalPathOnly(t *testing.T) {
	dir := t.TempDir()
	localPath := WorkerYAMLPath(filepath.Join(dir, ".worktrees", "dev-001"))
	localCfg := &WorkerConfig{WorkerID: "dev-001", Role: "local-role"}
	if err := localCfg.Save(localPath); err != nil {
		t.Fatalf("save local config: %v", err)
	}
	loaded, path, err := LoadWorkerConfigByID(dir, ".worktrees", "dev-001")
	if err != nil {
		t.Fatalf("LoadWorkerConfigByID: %v", err)
	}
	if loaded.Role != "local-role" {
		t.Fatalf("Role = %q, want local-role", loaded.Role)
	}
	if path != localPath {
		t.Fatalf("path = %q, want %q", path, localPath)
	}
}

func TestWorkerConfigWritePathPrefersLocalWhenWorktreeExists(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, ".worktrees", "dev-001")
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	if got := WorkerConfigWritePath(dir, ".worktrees", "dev-001"); got != WorkerYAMLPath(wtPath) {
		t.Fatalf("WorkerConfigWritePath = %q", got)
	}
}

func TestMainSessionConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := MainSessionYAMLPath(dir)

	original := &MainSessionConfig{
		Backend:   "wezterm",
		PaneID:    "200",
		UpdatedAt: "2026-03-19T10:00:00Z",
	}

	if err := original.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadMainSessionConfig(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Backend != original.Backend {
		t.Errorf("Backend = %q, want %q", loaded.Backend, original.Backend)
	}
	if loaded.PaneID != original.PaneID {
		t.Errorf("PaneID = %q, want %q", loaded.PaneID, original.PaneID)
	}
}

func TestLoadMainSessionConfigNotFound(t *testing.T) {
	_, err := LoadMainSessionConfig("/nonexistent/main-session.yaml")
	if err == nil {
		t.Fatal("expected error for missing main session config")
	}
}
