package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestWorkerOpenCmdRejectsPositionalProviderSyntax(t *testing.T) {
	cmd := newWorkerOpenCmd()
	cmd.SetArgs([]string{"backend-001", "codex"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected positional provider syntax to fail")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s), received 2") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunWorkerOpenPersistsExplicitOverrides(t *testing.T) {
	skillInstaller = func(_, _, _, _, _ string, _ bool) error { return nil }
	workerShellInitDelay = 0
	t.Cleanup(func() {
		skillInstaller = internal.InstallSkillsForWorkerFromPath
		workerShellInitDelay = 2 * time.Second
	})

	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "pane-3",
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

	wtPath := filepath.Join(dir, ".worktrees", "backend-001")
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	cfg := &internal.WorkerConfig{
		WorkerID:     "backend-001",
		Role:         "backend",
		Provider:     "claude",
		DefaultModel: "sonnet",
		PaneID:       "",
	}
	if err := cfg.Save(internal.WorkerConfigPath(dir, "backend-001")); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	if err := app.RunWorkerOpen("backend-001", "codex", "gpt-5", false, true, true); err != nil {
		t.Fatalf("RunWorkerOpen: %v", err)
	}

	reloaded, err := internal.LoadWorkerConfig(internal.WorkerConfigPath(dir, "backend-001"))
	if err != nil {
		t.Fatalf("LoadWorkerConfig: %v", err)
	}
	if reloaded.Provider != "codex" {
		t.Fatalf("Provider = %q, want codex", reloaded.Provider)
	}
	if reloaded.DefaultModel != "gpt-5" {
		t.Fatalf("DefaultModel = %q, want gpt-5", reloaded.DefaultModel)
	}
	if len(mock.SentTexts) == 0 {
		t.Fatal("expected launch command to be sent")
	}
	if got := mock.SentTexts[0]; got != "codex --dangerously-bypass-approvals-and-sandbox --model gpt-5" {
		t.Fatalf("launch command = %q", got)
	}
}

func TestRunWorkerOpenWithoutFlagsUsesPersistedConfig(t *testing.T) {
	skillInstaller = func(_, _, _, _, _ string, _ bool) error { return nil }
	workerShellInitDelay = 0
	t.Cleanup(func() {
		skillInstaller = internal.InstallSkillsForWorkerFromPath
		workerShellInitDelay = 2 * time.Second
	})

	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "pane-4",
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

	wtPath := filepath.Join(dir, ".worktrees", "backend-001")
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	cfg := &internal.WorkerConfig{
		WorkerID:     "backend-001",
		Role:         "backend",
		Provider:     "codex",
		DefaultModel: "gpt-5",
	}
	if err := cfg.Save(internal.WorkerConfigPath(dir, "backend-001")); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	if err := app.RunWorkerOpen("backend-001", "", "", false, false, false); err != nil {
		t.Fatalf("RunWorkerOpen: %v", err)
	}

	reloaded, err := internal.LoadWorkerConfig(internal.WorkerConfigPath(dir, "backend-001"))
	if err != nil {
		t.Fatalf("LoadWorkerConfig: %v", err)
	}
	if reloaded.Provider != "codex" {
		t.Fatalf("Provider = %q, want codex", reloaded.Provider)
	}
	if reloaded.DefaultModel != "gpt-5" {
		t.Fatalf("DefaultModel = %q, want gpt-5", reloaded.DefaultModel)
	}
	if len(mock.SentTexts) == 0 {
		t.Fatal("expected launch command to be sent")
	}
	if got := mock.SentTexts[0]; got != "codex --dangerously-bypass-approvals-and-sandbox --model gpt-5" {
		t.Fatalf("launch command = %q", got)
	}
}

func TestRunWorkerOpenCompatProviderFallbackDoesNotPersist(t *testing.T) {
	skillInstaller = func(_, _, _, _, _ string, _ bool) error { return nil }
	workerShellInitDelay = 0
	t.Cleanup(func() {
		skillInstaller = internal.InstallSkillsForWorkerFromPath
		workerShellInitDelay = 2 * time.Second
	})

	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "pane-5",
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

	wtPath := filepath.Join(dir, ".worktrees", "backend-001")
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	cfg := &internal.WorkerConfig{
		WorkerID: "backend-001",
		Role:     "backend",
	}
	if err := cfg.Save(internal.WorkerConfigPath(dir, "backend-001")); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	if err := app.RunWorkerOpen("backend-001", "", "", false, false, false); err != nil {
		t.Fatalf("RunWorkerOpen: %v", err)
	}

	reloaded, err := internal.LoadWorkerConfig(internal.WorkerConfigPath(dir, "backend-001"))
	if err != nil {
		t.Fatalf("LoadWorkerConfig: %v", err)
	}
	if reloaded.Provider != "" {
		t.Fatalf("Provider = %q, want empty", reloaded.Provider)
	}
	if len(mock.SentTexts) == 0 {
		t.Fatal("expected launch command to be sent")
	}
	if got := mock.SentTexts[0]; got != "claude --dangerously-skip-permissions" {
		t.Fatalf("launch command = %q", got)
	}
}

func TestRunWorkerOpenBootstrapsDeferredWorktree(t *testing.T) {
	skillInstaller = func(_, _, _, _, _ string, _ bool) error { return nil }
	taskSetup = func(string) error { return nil }
	workerShellInitDelay = 0
	t.Cleanup(func() {
		skillInstaller = internal.InstallSkillsForWorkerFromPath
		taskSetup = defaultTaskSetup
		workerShellInitDelay = 2 * time.Second
	})

	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "pane-6",
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
	if err := os.WriteFile(filepath.Join(roleDir, "system.md"), []byte("# backend\n"), 0644); err != nil {
		t.Fatalf("write system.md: %v", err)
	}

	falseValue := false
	cfg := &internal.WorkerConfig{
		WorkerID:        "backend-001",
		Role:            "backend",
		Provider:        "claude",
		WorktreeCreated: &falseValue,
	}
	if err := cfg.Save(internal.WorkerConfigPath(dir, "backend-001")); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	wtPath := filepath.Join(dir, ".worktrees", "backend-001")
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Fatalf("worktree should not exist before open, err=%v", err)
	}

	if err := app.RunWorkerOpen("backend-001", "", "", false, false, false); err != nil {
		t.Fatalf("RunWorkerOpen: %v", err)
	}

	if _, err := os.Stat(wtPath); err != nil {
		t.Fatalf("expected bootstrapped worktree: %v", err)
	}
	reloaded, err := internal.LoadWorkerConfig(internal.WorkerConfigPath(dir, "backend-001"))
	if err != nil {
		t.Fatalf("LoadWorkerConfig: %v", err)
	}
	if !reloaded.IsWorktreeCreated() {
		t.Fatal("WorktreeCreated should be true after bootstrap")
	}
	if len(mock.SentTexts) == 0 {
		t.Fatal("expected launch command to be sent")
	}
	if got := mock.SentTexts[0]; got != "claude --dangerously-skip-permissions" {
		t.Fatalf("launch command = %q", got)
	}
}

func TestRunWorkerOpenFallsBackToLegacyConfig(t *testing.T) {
	skillInstaller = func(_, _, _, _, _ string, _ bool) error { return nil }
	workerShellInitDelay = 0
	t.Cleanup(func() {
		skillInstaller = internal.InstallSkillsForWorkerFromPath
		workerShellInitDelay = 2 * time.Second
	})

	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "pane-7",
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
	if err := os.WriteFile(filepath.Join(roleDir, "system.md"), []byte("# backend\n"), 0644); err != nil {
		t.Fatalf("write system.md: %v", err)
	}

	wtPath := filepath.Join(dir, ".worktrees", "backend-001")
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	cfg := &internal.WorkerConfig{
		WorkerID:     "backend-001",
		Role:         "backend",
		Provider:     "claude",
		DefaultModel: "sonnet",
	}
	if err := cfg.Save(internal.WorkerYAMLPath(wtPath)); err != nil {
		t.Fatalf("save legacy worker config: %v", err)
	}

	if err := app.RunWorkerOpen("backend-001", "", "", false, false, false); err != nil {
		t.Fatalf("RunWorkerOpen: %v", err)
	}
	if len(mock.SentTexts) == 0 {
		t.Fatal("expected launch command to be sent")
	}
	if got := mock.SentTexts[0]; got != "claude --dangerously-skip-permissions --model sonnet" {
		t.Fatalf("launch command = %q", got)
	}
}
