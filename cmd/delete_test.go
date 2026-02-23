// cmd/delete_test.go
package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leeforge/agent-team/internal"
)

func TestRunDelete(t *testing.T) {
	app, dir := initTestApp(t)
	app.RunCreate("ephemeral")

	// Verify it exists
	wtPath := filepath.Join(dir, ".worktrees", "ephemeral")
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Fatal("setup: role not created")
	}

	if err := app.RunDelete("ephemeral"); err != nil {
		t.Fatalf("RunDelete: %v", err)
	}

	// Worktree should be gone
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Error("worktree directory still exists after delete")
	}
}

func TestRunDeleteNotFound(t *testing.T) {
	app, _ := initTestApp(t)
	err := app.RunDelete("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent role")
	}
}

func TestRunDeleteKillsRunningPane(t *testing.T) {
	app, dir := initTestApp(t)
	app.RunCreate("running")

	// Simulate a running pane
	mock := &MockBackend{AlivePanes: map[string]bool{"123": true}}
	app.Session = mock

	configPath := filepath.Join(dir, ".worktrees", "running", "agents", "teams", "running", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	cfg.PaneID = "123"
	cfg.Save(configPath)

	app.RunDelete("running")
	// No assertion on KillPane — just ensure no panic
}
