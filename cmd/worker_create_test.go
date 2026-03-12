package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestWorkerCreateCmdRejectsPositionalProviderSyntax(t *testing.T) {
	cmd := newWorkerCreateCmd()
	cmd.SetArgs([]string{"backend", "codex"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected positional provider syntax to fail")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s), received 2") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunWorkerCreateDefaultsProviderAndPersistsModel(t *testing.T) {
	taskSetup = func(string) error { return nil }
	skillInstaller = func(_, _, _, _, _ string, _ bool) error { return nil }
	workerShellInitDelay = 0
	t.Cleanup(func() {
		taskSetup = defaultTaskSetup
		skillInstaller = internal.InstallSkillsForWorkerFromPath
		workerShellInitDelay = 2 * time.Second
	})

	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "pane-1",
		AlivePanes: map[string]bool{},
	}
	app.Session = mock

	roleDir := filepath.Join(dir, ".agents", "teams", "backend")
	if err := os.MkdirAll(roleDir, 0755); err != nil {
		t.Fatalf("mkdir role dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# backend\n"), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	if err := app.RunWorkerCreate("backend", "", "gpt-5", false, false); err != nil {
		t.Fatalf("RunWorkerCreate: %v", err)
	}

	cfg, err := internal.LoadWorkerConfig(filepath.Join(dir, ".worktrees", "backend-001", "worker.yaml"))
	if err != nil {
		t.Fatalf("LoadWorkerConfig: %v", err)
	}
	if cfg.Provider != "claude" {
		t.Fatalf("Provider = %q, want claude", cfg.Provider)
	}
	if cfg.DefaultModel != "gpt-5" {
		t.Fatalf("DefaultModel = %q, want gpt-5", cfg.DefaultModel)
	}
	if len(mock.SentTexts) == 0 {
		t.Fatal("expected launch command to be sent")
	}
	if got := mock.SentTexts[0]; got != "claude --dangerously-skip-permissions --model gpt-5" {
		t.Fatalf("launch command = %q", got)
	}
}

func TestRunWorkerCreatePersistsExplicitProvider(t *testing.T) {
	taskSetup = func(string) error { return nil }
	skillInstaller = func(_, _, _, _, _ string, _ bool) error { return nil }
	workerShellInitDelay = 0
	t.Cleanup(func() {
		taskSetup = defaultTaskSetup
		skillInstaller = internal.InstallSkillsForWorkerFromPath
		workerShellInitDelay = 2 * time.Second
	})

	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "pane-2",
		AlivePanes: map[string]bool{},
	}
	app.Session = mock

	roleDir := filepath.Join(dir, ".agents", "teams", "backend")
	if err := os.MkdirAll(roleDir, 0755); err != nil {
		t.Fatalf("mkdir role dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# backend\n"), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	if err := app.RunWorkerCreate("backend", "codex", "", false, false); err != nil {
		t.Fatalf("RunWorkerCreate: %v", err)
	}

	cfg, err := internal.LoadWorkerConfig(filepath.Join(dir, ".worktrees", "backend-001", "worker.yaml"))
	if err != nil {
		t.Fatalf("LoadWorkerConfig: %v", err)
	}
	if cfg.Provider != "codex" {
		t.Fatalf("Provider = %q, want codex", cfg.Provider)
	}
}
