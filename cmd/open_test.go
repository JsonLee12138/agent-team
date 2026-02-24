// cmd/open_test.go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunOpen(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &MockBackend{SpawnedID: "77"}
	app.Session = mock
	app.RunCreate("dev")

	if err := app.RunOpen("dev", "", ""); err != nil {
		t.Fatalf("RunOpen: %v", err)
	}

	// pane_id should be saved
	configPath := filepath.Join(dir, ".worktrees", "dev", "agents", "teams", "dev", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	if cfg.PaneID != "77" {
		t.Errorf("PaneID = %q, want 77", cfg.PaneID)
	}

	// CLAUDE.md should exist
	claudeMD := filepath.Join(dir, ".worktrees", "dev", "CLAUDE.md")
	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("CLAUDE.md not found: %v", err)
	}
	if !strings.Contains(string(data), "team/dev") {
		t.Error("CLAUDE.md missing worktree context")
	}

	// launch command should be sent
	if len(mock.SentTexts) == 0 {
		t.Error("no command sent to pane")
	}
	if !strings.Contains(mock.SentTexts[0], "claude") {
		t.Errorf("sent text = %q, want claude launch command", mock.SentTexts[0])
	}
}

func TestRunOpenAlreadyRunning(t *testing.T) {
	app, dir := initTestApp(t)
	app.RunCreate("running")

	// Set pane_id and mark alive
	configPath := filepath.Join(dir, ".worktrees", "running", "agents", "teams", "running", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	cfg.PaneID = "55"
	cfg.Save(configPath)

	mock := &MockBackend{AlivePanes: map[string]bool{"55": true}}
	app.Session = mock

	// Should not error, just print message
	err := app.RunOpen("running", "", "")
	if err != nil {
		t.Fatalf("RunOpen on running role should not error: %v", err)
	}
}

func TestRunOpenWithProvider(t *testing.T) {
	app, _ := initTestApp(t)
	mock := &MockBackend{SpawnedID: "88"}
	app.Session = mock
	app.RunCreate("test")

	app.RunOpen("test", "codex", "gpt-5")

	if len(mock.SentTexts) == 0 {
		t.Fatal("no command sent")
	}
	if !strings.Contains(mock.SentTexts[0], "codex") {
		t.Errorf("sent = %q, want codex command", mock.SentTexts[0])
	}
	if !strings.Contains(mock.SentTexts[0], "--model gpt-5") {
		t.Errorf("sent = %q, want --model gpt-5", mock.SentTexts[0])
	}
}
