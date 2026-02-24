// cmd/assign_test.go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunAssign(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "99",
		AlivePanes: map[string]bool{"99": true},
	}
	app.Session = mock
	app.RunCreate("worker")

	// Set pane as running
	configPath := filepath.Join(dir, ".worktrees", "worker", "agents", "teams", "worker", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	cfg.PaneID = "99"
	cfg.Save(configPath)

	err := app.RunAssign("worker", "Fix the login bug", "", "")
	if err != nil {
		t.Fatalf("RunAssign: %v", err)
	}

	// Task file should exist in pending
	pendingDir := filepath.Join(dir, ".worktrees", "worker", "agents", "teams", "worker", "tasks", "pending")
	entries, _ := os.ReadDir(pendingDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 task file, got %d", len(entries))
	}

	data, _ := os.ReadFile(filepath.Join(pendingDir, entries[0].Name()))
	if !strings.Contains(string(data), "Fix the login bug") {
		t.Error("task file missing task description")
	}

	// Notification should be sent to pane
	if len(mock.SentTexts) == 0 {
		t.Error("no notification sent to pane")
	}
}

func TestRunAssignNotFound(t *testing.T) {
	app, _ := initTestApp(t)
	err := app.RunAssign("ghost", "task", "", "")
	if err == nil {
		t.Error("expected error for nonexistent role")
	}
}
