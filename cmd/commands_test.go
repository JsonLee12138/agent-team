// cmd/commands_test.go
package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leeforge/agent-team/internal"
)

// --- reply ---

func TestRunReply(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "50",
		AlivePanes: map[string]bool{"50": true},
	}
	app.Session = mock
	app.RunCreate("dev")

	configPath := filepath.Join(dir, ".worktrees", "dev", "agents", "teams", "dev", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	cfg.PaneID = "50"
	cfg.Save(configPath)

	err := app.RunReply("dev", "Use the factory pattern")
	if err != nil {
		t.Fatalf("RunReply: %v", err)
	}

	if len(mock.SentTexts) == 0 {
		t.Fatal("no text sent")
	}
	if !strings.Contains(mock.SentTexts[0], "[Main Controller Reply]") {
		t.Errorf("sent = %q, want [Main Controller Reply] prefix", mock.SentTexts[0])
	}
}

func TestRunReplyOffline(t *testing.T) {
	app, _ := initTestApp(t)
	app.RunCreate("offline")
	err := app.RunReply("offline", "hello")
	if err == nil {
		t.Error("expected error when role is offline")
	}
}

// --- status ---

func TestRunStatus(t *testing.T) {
	app, _ := initTestApp(t)
	app.RunCreate("alpha")
	app.RunCreate("beta")

	// Should not error
	err := app.RunStatus()
	if err != nil {
		t.Fatalf("RunStatus: %v", err)
	}
}

// --- merge ---

func TestRunMerge(t *testing.T) {
	app, dir := initTestApp(t)
	app.RunCreate("feature")

	// Make a commit in the worktree so merge has something
	wtPath := filepath.Join(dir, ".worktrees", "feature")
	commitFile := filepath.Join(wtPath, "test.txt")
	os.WriteFile(commitFile, []byte("test"), 0644)
	exec.Command("git", "-C", wtPath, "add", ".").Run()
	exec.Command("git", "-C", wtPath, "commit", "-m", "test commit").Run()

	err := app.RunMerge("feature")
	if err != nil {
		t.Fatalf("RunMerge: %v", err)
	}
}

func TestRunMergeNotFound(t *testing.T) {
	app, _ := initTestApp(t)
	err := app.RunMerge("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent role")
	}
}
