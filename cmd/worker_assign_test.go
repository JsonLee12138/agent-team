package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunWorkerAssignSkipsRebaseForUnmaterializedWorker(t *testing.T) {
	skillInstaller = func(_, _, _, _, _ string, _ bool) error { return nil }
	taskSetup = func(string) error { return nil }
	workerShellInitDelay = 0
	t.Cleanup(func() {
		skillInstaller = internal.InstallSkillsForWorkerFromPath
		taskSetup = defaultTaskSetup
		workerShellInitDelay = 2 * time.Second
	})

	app, dir := initTestApp(t)
	mock := &MockBackend{SpawnedID: "pane-1", AlivePanes: map[string]bool{}}
	app.Session = mock

	roleDir := filepath.Join(dir, ".agents", "teams", "backend")
	if err := os.MkdirAll(roleDir, 0755); err != nil {
		t.Fatalf("mkdir role dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# backend\n"), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	falseValue := false
	cfg := &internal.WorkerConfig{WorkerID: "backend-001", Role: "backend", Provider: "claude", WorktreeCreated: &falseValue}
	if err := cfg.Save(internal.WorkerConfigPath(dir, "backend-001")); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	if err := app.RunWorkerAssign("backend-001", "Implement feature", "", "", "", "", false); err != nil {
		t.Fatalf("RunWorkerAssign: %v", err)
	}
	if len(mock.SentTexts) < 2 {
		t.Fatalf("sent texts = %#v, want launch + assignment", mock.SentTexts)
	}
	if got := mock.SentTexts[0]; got != "claude --dangerously-skip-permissions" {
		t.Fatalf("launch command = %q", got)
	}
	if _, err := os.Stat(filepath.Join(dir, ".worktrees", "backend-001")); err != nil {
		t.Fatalf("expected worktree to exist after assign bootstrap: %v", err)
	}
}

func TestRunWorkerAssignRebasesIdleWorker(t *testing.T) {
	skillInstaller = func(_, _, _, _, _ string, _ bool) error { return nil }
	workerShellInitDelay = 0
	t.Cleanup(func() {
		skillInstaller = internal.InstallSkillsForWorkerFromPath
		workerShellInitDelay = 2 * time.Second
	})

	app, dir := initTestApp(t)
	mock := &MockBackend{SpawnedID: "pane-2", AlivePanes: map[string]bool{}}
	app.Session = mock

	roleDir := filepath.Join(dir, ".agents", "teams", "backend")
	os.MkdirAll(roleDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# backend\n"), 0644)

	wtPath := filepath.Join(dir, ".worktrees", "backend-001")
	if err := app.Git.WorktreeAdd(wtPath, "team/backend-001"); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}
	defer app.Git.WorktreeRemove(wtPath)
	trueValue := true
	cfg := &internal.WorkerConfig{WorkerID: "backend-001", Role: "backend", Provider: "claude", WorktreeCreated: &trueValue}
	if err := cfg.Save(internal.WorkerConfigPath(dir, "backend-001")); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	run := func(cwd string, args ...string) {
		cmd, err := internal.NewGitClient(cwd)
		if err != nil || cmd == nil {
			t.Fatalf("NewGitClient(%s): %v", cwd, err)
		}
		_ = cmd
	}
	_ = run
	if err := os.WriteFile(filepath.Join(dir, "probe.txt"), []byte("main\n"), 0644); err != nil {
		t.Fatalf("write main file: %v", err)
	}

	if err := app.RunWorkerAssign("backend-001", "Implement feature", "", "", "", "", false); err != nil {
		t.Fatalf("RunWorkerAssign: %v", err)
	}
}

func TestRunWorkerAssignSkipsRebaseWhenActiveChangesExist(t *testing.T) {
	skillInstaller = func(_, _, _, _, _ string, _ bool) error { return nil }
	workerShellInitDelay = 0
	t.Cleanup(func() {
		skillInstaller = internal.InstallSkillsForWorkerFromPath
		workerShellInitDelay = 2 * time.Second
	})

	app, dir := initTestApp(t)
	mock := &MockBackend{SpawnedID: "pane-3", AlivePanes: map[string]bool{}}
	app.Session = mock

	roleDir := filepath.Join(dir, ".agents", "teams", "backend")
	os.MkdirAll(roleDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# backend\n"), 0644)

	wtPath := filepath.Join(dir, ".worktrees", "backend-001")
	if err := app.Git.WorktreeAdd(wtPath, "team/backend-001"); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}
	defer app.Git.WorktreeRemove(wtPath)
	trueValue := true
	cfg := &internal.WorkerConfig{WorkerID: "backend-001", Role: "backend", Provider: "claude", WorktreeCreated: &trueValue}
	if err := cfg.Save(internal.WorkerConfigPath(dir, "backend-001")); err != nil {
		t.Fatalf("save worker config: %v", err)
	}
	if err := internal.InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir: %v", err)
	}
	if _, err := internal.CreateTaskChange(wtPath, "2026-03-19-existing", "existing", "", ""); err != nil {
		t.Fatalf("CreateTaskChange: %v", err)
	}

	if err := app.RunWorkerAssign("backend-001", "Implement feature", "", "", "", "", false); err != nil {
		t.Fatalf("RunWorkerAssign: %v", err)
	}
}
