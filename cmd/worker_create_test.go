package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stderr = w
	t.Cleanup(func() {
		os.Stderr = orig
	})

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	return string(out)
}

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
	t.Cleanup(func() {
		taskSetup = defaultTaskSetup
		skillInstaller = internal.InstallSkillsForWorkerFromPath
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

	if err := app.RunWorkerCreate("backend", "", "gpt-5", false); err != nil {
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
	// create no longer spawns a pane or sends launch commands
	if cfg.PaneID != "" {
		t.Fatalf("PaneID = %q, want empty (create should not open session)", cfg.PaneID)
	}
	if len(mock.SentTexts) != 0 {
		t.Fatalf("SentTexts = %v, want none (create should not launch provider)", mock.SentTexts)
	}
}

func TestRunWorkerCreatePersistsExplicitProvider(t *testing.T) {
	taskSetup = func(string) error { return nil }
	skillInstaller = func(_, _, _, _, _ string, _ bool) error { return nil }
	t.Cleanup(func() {
		taskSetup = defaultTaskSetup
		skillInstaller = internal.InstallSkillsForWorkerFromPath
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

	if err := app.RunWorkerCreate("backend", "codex", "", false); err != nil {
		t.Fatalf("RunWorkerCreate: %v", err)
	}

	cfg, err := internal.LoadWorkerConfig(filepath.Join(dir, ".worktrees", "backend-001", "worker.yaml"))
	if err != nil {
		t.Fatalf("LoadWorkerConfig: %v", err)
	}
	if cfg.Provider != "codex" {
		t.Fatalf("Provider = %q, want codex", cfg.Provider)
	}
	if cfg.PaneID != "" {
		t.Fatalf("PaneID = %q, want empty", cfg.PaneID)
	}
}

func TestRunWorkerCreateSkillsSyncWarningDoesNotFail(t *testing.T) {
	taskSetup = func(string) error { return nil }
	skillInstaller = func(_, _, _, _, _ string, _ bool) error {
		return fmt.Errorf("npm install failed")
	}
	t.Cleanup(func() {
		taskSetup = defaultTaskSetup
		skillInstaller = internal.InstallSkillsForWorkerFromPath
	})

	app, dir := initTestApp(t)
	app.Session = &MockBackend{AlivePanes: map[string]bool{}}

	roleDir := filepath.Join(dir, ".agents", "teams", "backend")
	os.MkdirAll(roleDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# backend\n"), 0644)

	// create should succeed despite skill sync failure
	if err := app.RunWorkerCreate("backend", "", "", false); err != nil {
		t.Fatalf("RunWorkerCreate should succeed on skill sync failure: %v", err)
	}
}

func TestRunWorkerCreateWarnsWhenBuildRulesHashIsStale(t *testing.T) {
	taskSetup = func(string) error { return nil }
	skillInstaller = func(_, _, _, _, _ string, _ bool) error { return nil }
	t.Cleanup(func() {
		taskSetup = defaultTaskSetup
		skillInstaller = internal.InstallSkillsForWorkerFromPath
	})

	app, dir := initTestApp(t)
	app.Session = &MockBackend{AlivePanes: map[string]bool{}}

	roleDir := filepath.Join(dir, ".agents", "teams", "backend")
	if err := os.MkdirAll(roleDir, 0755); err != nil {
		t.Fatalf("mkdir role dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# backend\n"), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:\n\tgo build ./...\n"), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	if _, err := internal.RebuildBuildRules(dir); err != nil {
		t.Fatalf("RebuildBuildRules: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:\n\tgo build ./...\n\ntest:\n\tgo test ./...\n"), 0644); err != nil {
		t.Fatalf("update Makefile: %v", err)
	}

	stderr := captureStderr(t, func() {
		if err := app.RunWorkerCreate("backend", "", "", false); err != nil {
			t.Fatalf("RunWorkerCreate: %v", err)
		}
	})

	if !strings.Contains(stderr, "build scripts changed since the last rules rebuild") {
		t.Fatalf("stderr = %q, want stale build rules warning", stderr)
	}
	if !strings.Contains(stderr, "agent-team rules sync --rebuild") {
		t.Fatalf("stderr = %q, want rebuild hint", stderr)
	}
}
