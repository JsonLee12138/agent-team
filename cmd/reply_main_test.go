// cmd/reply_main_test.go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunReplyMain(t *testing.T) {
	openSpecSetup = func(wtPath string) error { return nil }
	defer func() { openSpecSetup = defaultOpenSpecSetup }()

	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "50",
		AlivePanes: map[string]bool{"50": true, "99": true},
	}
	app.Session = mock
	app.RunCreate("dev")

	// Set controller pane ID in config
	configPath := filepath.Join(dir, ".worktrees", "dev", "agents", "teams", "dev", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	cfg.PaneID = "50"
	cfg.ControllerPaneID = "99"
	cfg.Save(configPath)

	// Change to worktree directory (reply-main uses os.Getwd())
	origDir, _ := os.Getwd()
	wtPath := filepath.Join(dir, ".worktrees", "dev")
	os.Chdir(wtPath)
	defer os.Chdir(origDir)

	err := app.RunReplyMain("What database should I use?")
	if err != nil {
		t.Fatalf("RunReplyMain: %v", err)
	}

	if len(mock.SentTexts) == 0 {
		t.Fatal("no text sent")
	}
	if !strings.Contains(mock.SentTexts[0], "[Role: dev]") {
		t.Errorf("sent = %q, want [Role: dev] prefix", mock.SentTexts[0])
	}
	if !strings.Contains(mock.SentTexts[0], "What database should I use?") {
		t.Errorf("sent = %q, want message content", mock.SentTexts[0])
	}
}

func TestRunReplyMainNoController(t *testing.T) {
	openSpecSetup = func(wtPath string) error { return nil }
	defer func() { openSpecSetup = defaultOpenSpecSetup }()

	app, dir := initTestApp(t)
	app.RunCreate("solo")

	// No controller_pane_id set
	origDir, _ := os.Getwd()
	wtPath := filepath.Join(dir, ".worktrees", "solo")
	os.Chdir(wtPath)
	defer os.Chdir(origDir)

	err := app.RunReplyMain("hello")
	if err == nil {
		t.Error("expected error when no controller pane ID")
	}
}

func TestRunReplyMainControllerOffline(t *testing.T) {
	openSpecSetup = func(wtPath string) error { return nil }
	defer func() { openSpecSetup = defaultOpenSpecSetup }()

	app, dir := initTestApp(t)
	mock := &MockBackend{
		AlivePanes: map[string]bool{}, // controller pane not alive
	}
	app.Session = mock
	app.RunCreate("offline")

	configPath := filepath.Join(dir, ".worktrees", "offline", "agents", "teams", "offline", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	cfg.ControllerPaneID = "99"
	cfg.Save(configPath)

	origDir, _ := os.Getwd()
	wtPath := filepath.Join(dir, ".worktrees", "offline")
	os.Chdir(wtPath)
	defer os.Chdir(origDir)

	err := app.RunReplyMain("hello")
	if err == nil {
		t.Error("expected error when controller is offline")
	}
}
