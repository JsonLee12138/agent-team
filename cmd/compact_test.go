package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunCompactFromWorkerWorktreeUsesWorkerPane(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &MockBackend{AlivePanes: map[string]bool{"50": true}}
	app.Session = mock

	wtPath := filepath.Join(dir, ".worktrees", "dev-001")
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	cfg := &internal.WorkerConfig{WorkerID: "dev-001", Role: "dev", Provider: "claude", PaneID: "50", ControllerPaneID: "99"}
	if err := cfg.Save(internal.WorkerConfigPath(dir, "dev-001")); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	origResolve := resolveGitTopLevel
	resolveGitTopLevel = func() (string, error) { return wtPath, nil }
	defer func() { resolveGitTopLevel = origResolve }()

	if err := app.RunCompact("", "", "", ""); err != nil {
		t.Fatalf("RunCompact: %v", err)
	}
	if len(mock.SentTexts) != 1 || mock.SentTexts[0] != "/compact" {
		t.Fatalf("sent texts = %#v, want [/compact]", mock.SentTexts)
	}
}

func TestRunCompactUsesExplicitPaneID(t *testing.T) {
	app, _ := initTestApp(t)
	mock := &MockBackend{AlivePanes: map[string]bool{"123": true}}
	app.Session = mock

	origResolve := resolveGitTopLevel
	resolveGitTopLevel = func() (string, error) { return app.Git.Root(), nil }
	defer func() { resolveGitTopLevel = origResolve }()

	if err := app.RunCompact("123", "", "", "Focus on current task"); err != nil {
		t.Fatalf("RunCompact: %v", err)
	}
	if len(mock.SentTexts) != 1 || mock.SentTexts[0] != "/compact Focus on current task" {
		t.Fatalf("sent texts = %#v", mock.SentTexts)
	}
}

func TestRunCompactWithWorkerFlagUsesWorkerPane(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &MockBackend{AlivePanes: map[string]bool{"50": true}}
	app.Session = mock

	wtPath := filepath.Join(dir, ".worktrees", "dev-001")
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	cfg := &internal.WorkerConfig{WorkerID: "dev-001", Role: "dev", Provider: "claude", PaneID: "50"}
	if err := cfg.Save(internal.WorkerConfigPath(dir, "dev-001")); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	origResolve := resolveGitTopLevel
	resolveGitTopLevel = func() (string, error) { return dir, nil }
	defer func() { resolveGitTopLevel = origResolve }()

	if err := app.RunCompact("", "dev-001", "", ""); err != nil {
		t.Fatalf("RunCompact: %v", err)
	}
	if len(mock.SentTexts) != 1 || mock.SentTexts[0] != "/compact" {
		t.Fatalf("sent texts = %#v", mock.SentTexts)
	}
}

func TestRunCompactFromWorkerWorktreeToMainUsesControllerPane(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &MockBackend{AlivePanes: map[string]bool{"99": true}}
	app.Session = mock

	wtPath := filepath.Join(dir, ".worktrees", "dev-001")
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	cfg := &internal.WorkerConfig{WorkerID: "dev-001", Role: "dev", Provider: "claude", PaneID: "50", ControllerPaneID: "99"}
	if err := cfg.Save(internal.WorkerConfigPath(dir, "dev-001")); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	origResolve := resolveGitTopLevel
	resolveGitTopLevel = func() (string, error) { return wtPath, nil }
	defer func() { resolveGitTopLevel = origResolve }()

	if err := app.RunCompact("", "", "main", ""); err != nil {
		t.Fatalf("RunCompact: %v", err)
	}
	if len(mock.SentTexts) != 1 || mock.SentTexts[0] != "/compact" {
		t.Fatalf("sent texts = %#v", mock.SentTexts)
	}
}

func TestRunCompactFromRootToMainUsesProjectConfig(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &MockBackend{AlivePanes: map[string]bool{"200": true}}
	app.Session = mock

	cfg := &internal.MainSessionConfig{PaneID: "200", Backend: "wezterm"}
	if err := cfg.Save(internal.MainSessionYAMLPath(dir)); err != nil {
		t.Fatalf("save main session config: %v", err)
	}

	origResolve := resolveGitTopLevel
	resolveGitTopLevel = func() (string, error) { return dir, nil }
	defer func() { resolveGitTopLevel = origResolve }()

	if err := app.RunCompact("", "", "main", ""); err != nil {
		t.Fatalf("RunCompact: %v", err)
	}
	if len(mock.SentTexts) != 1 || mock.SentTexts[0] != "/compact" {
		t.Fatalf("sent texts = %#v", mock.SentTexts)
	}
}

func TestRunCompactFromRootToMainFallsBackToEnvAndSavesConfig(t *testing.T) {
	app, dir := initTestApp(t)
	t.Setenv("WEZTERM_PANE", "321")
	mock := &MockBackend{AlivePanes: map[string]bool{"321": true}}
	app.Session = mock

	origResolve := resolveGitTopLevel
	resolveGitTopLevel = func() (string, error) { return dir, nil }
	defer func() { resolveGitTopLevel = origResolve }()

	if err := app.RunCompact("", "", "main", ""); err != nil {
		t.Fatalf("RunCompact: %v", err)
	}

	saved, err := internal.LoadMainSessionConfig(internal.MainSessionYAMLPath(dir))
	if err != nil {
		t.Fatalf("LoadMainSessionConfig: %v", err)
	}
	if saved.PaneID != "321" || saved.Backend != "wezterm" {
		t.Fatalf("saved config = %#v", saved)
	}
}

func TestRunCompactReturnsErrorWhenPaneNotAlive(t *testing.T) {
	app, _ := initTestApp(t)
	mock := &MockBackend{AlivePanes: map[string]bool{}}
	app.Session = mock

	origResolve := resolveGitTopLevel
	resolveGitTopLevel = func() (string, error) { return app.Git.Root(), nil }
	defer func() { resolveGitTopLevel = origResolve }()

	err := app.RunCompact("123", "", "", "")
	if err == nil || !strings.Contains(err.Error(), "target pane 123 is not running") {
		t.Fatalf("err = %v, want pane not running", err)
	}
}

func TestRunCompactReturnsErrorWhenTargetMissing(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &MockBackend{AlivePanes: map[string]bool{}}
	app.Session = mock

	origResolve := resolveGitTopLevel
	resolveGitTopLevel = func() (string, error) { return dir, nil }
	defer func() { resolveGitTopLevel = origResolve }()

	err := app.RunCompact("", "", "", "")
	if err == nil || !strings.Contains(err.Error(), "must specify --worker <worker-id>") {
		t.Fatalf("err = %v, want missing target error", err)
	}
}

func TestRunCompactToMainReturnsErrorWhenNoFallbackEnv(t *testing.T) {
	t.Setenv("WEZTERM_PANE", "")
	t.Setenv("TMUX_PANE", "")
	app, dir := initTestApp(t)
	mock := &MockBackend{AlivePanes: map[string]bool{}}
	app.Session = mock

	origResolve := resolveGitTopLevel
	resolveGitTopLevel = func() (string, error) { return dir, nil }
	defer func() { resolveGitTopLevel = origResolve }()

	err := app.RunCompact("", "", "main", "")
	if err == nil || !strings.Contains(err.Error(), "no project main pane recorded") {
		t.Fatalf("err = %v, want main pane fallback error", err)
	}
}

func TestDetectCurrentPaneFromEnv(t *testing.T) {
	t.Run("prefers wezterm", func(t *testing.T) {
		t.Setenv("WEZTERM_PANE", "wez-1")
		t.Setenv("TMUX_PANE", "%9")
		paneID, backend := detectCurrentPaneFromEnv()
		if paneID != "wez-1" || backend != "wezterm" {
			t.Fatalf("got (%q, %q)", paneID, backend)
		}
	})

	t.Run("falls back to tmux", func(t *testing.T) {
		t.Setenv("WEZTERM_PANE", "")
		t.Setenv("TMUX_PANE", "%9")
		paneID, backend := detectCurrentPaneFromEnv()
		if paneID != "%9" || backend != "tmux" {
			t.Fatalf("got (%q, %q)", paneID, backend)
		}
	})
}
