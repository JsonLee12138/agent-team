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
	openSpecSetup = func(wtPath string) error { return nil }
	defer func() { openSpecSetup = defaultOpenSpecSetup }()

	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "99",
		AlivePanes: map[string]bool{"99": true},
	}
	app.Session = mock
	app.RunCreate("worker")

	// Manually create openspec/changes/ directory (normally done by openspec init)
	wtPath := filepath.Join(dir, ".worktrees", "worker")
	os.MkdirAll(filepath.Join(wtPath, "openspec", "changes"), 0755)

	// Set pane as running
	configPath := filepath.Join(wtPath, "agents", "teams", "worker", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	cfg.PaneID = "99"
	cfg.Save(configPath)

	err := app.RunAssign("worker", "Fix the login bug", "", "", "")
	if err != nil {
		t.Fatalf("RunAssign: %v", err)
	}

	// Change directory should exist under openspec/changes/
	changesDir := filepath.Join(wtPath, "openspec", "changes")
	entries, _ := os.ReadDir(changesDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 change directory, got %d", len(entries))
	}

	// proposal.md should exist (empty since no --proposal)
	proposalPath := filepath.Join(changesDir, entries[0].Name(), "proposal.md")
	if _, err := os.Stat(proposalPath); os.IsNotExist(err) {
		t.Error("proposal.md not created")
	}

	// .openspec.yaml should exist
	metaPath := filepath.Join(changesDir, entries[0].Name(), ".openspec.yaml")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error(".openspec.yaml not created")
	}

	// Notification should be sent to pane
	if len(mock.SentTexts) == 0 {
		t.Error("no notification sent to pane")
	}
	if !strings.Contains(mock.SentTexts[0], "[New Change Assigned]") {
		t.Errorf("notification = %q, want [New Change Assigned] prefix", mock.SentTexts[0])
	}
	if !strings.Contains(mock.SentTexts[0], "/opsx:continue") {
		t.Errorf("notification should mention /opsx:continue")
	}
}

func TestRunAssignWithProposal(t *testing.T) {
	openSpecSetup = func(wtPath string) error { return nil }
	defer func() { openSpecSetup = defaultOpenSpecSetup }()

	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "99",
		AlivePanes: map[string]bool{"99": true},
	}
	app.Session = mock
	app.RunCreate("worker")

	wtPath := filepath.Join(dir, ".worktrees", "worker")
	os.MkdirAll(filepath.Join(wtPath, "openspec", "changes"), 0755)

	configPath := filepath.Join(wtPath, "agents", "teams", "worker", "config.yaml")
	cfg, _ := internal.LoadRoleConfig(configPath)
	cfg.PaneID = "99"
	cfg.Save(configPath)

	// Create a proposal file
	proposalFile := filepath.Join(dir, "proposal.md")
	proposalContent := "# Proposal\n\nFix the login flow with JWT."
	os.WriteFile(proposalFile, []byte(proposalContent), 0644)

	err := app.RunAssign("worker", "Fix the login bug", "", "", proposalFile)
	if err != nil {
		t.Fatalf("RunAssign: %v", err)
	}

	// Verify proposal content was written
	changesDir := filepath.Join(wtPath, "openspec", "changes")
	entries, _ := os.ReadDir(changesDir)
	proposalPath := filepath.Join(changesDir, entries[0].Name(), "proposal.md")
	data, _ := os.ReadFile(proposalPath)
	if string(data) != proposalContent {
		t.Errorf("proposal content = %q, want %q", string(data), proposalContent)
	}
}

func TestRunAssignNotFound(t *testing.T) {
	app, _ := initTestApp(t)
	err := app.RunAssign("ghost", "task", "", "", "")
	if err == nil {
		t.Error("expected error for nonexistent role")
	}
}
