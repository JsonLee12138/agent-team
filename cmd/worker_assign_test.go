package cmd

import (
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunWorkerAssignDelegatesToTaskAssignCompat(t *testing.T) {
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
	record, err := internal.CreateTaskPackage(dir, "Implement feature", "backend", "", time.Now().UTC())
	if err != nil {
		t.Fatalf("CreateTaskPackage: %v", err)
	}
	worktreeCreated := false
	cfg := &internal.WorkerConfig{WorkerID: "backend-001", Role: "backend", Provider: "claude", WorktreeCreated: &worktreeCreated}
	if err := cfg.Save(internal.WorkerConfigPath(dir, "backend-001")); err != nil {
		t.Fatalf("save worker config: %v", err)
	}
	if err := app.RunWorkerAssign("backend-001", record.TaskID, "", "", false); err != nil {
		t.Fatalf("RunWorkerAssign: %v", err)
	}
}
