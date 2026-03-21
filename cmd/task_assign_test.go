package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunTaskAssignCreatesWorkerAndSendsLightReminder(t *testing.T) {
	skillInstaller = func(_, _, _, _, _ string, _ bool) error { return nil }
	taskSetup = func(string) error { return nil }
	workerShellInitDelay = 0
	defer func() {
		skillInstaller = internal.InstallSkillsForWorkerFromPath
		taskSetup = defaultTaskSetup
		workerShellInitDelay = 2 * time.Second
	}()

	app, dir := initTestApp(t)
	mock := &MockBackend{SpawnedID: "pane-1", AlivePanes: map[string]bool{}}
	app.Session = mock
	roleDir := filepath.Join(dir, ".agents", "teams", "backend")
	refDir := filepath.Join(roleDir, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# backend\n"), 0644)
	os.WriteFile(filepath.Join(roleDir, "system.md"), []byte("# backend\n"), 0644)
	os.WriteFile(filepath.Join(refDir, "role.yaml"), []byte("name: backend\n"), 0644)
	record, err := internal.CreateTaskPackage(dir, "Implement feature", "backend", "", time.Now().UTC())
	if err != nil {
		t.Fatalf("CreateTaskPackage: %v", err)
	}
	if err := app.RunTaskAssign(record.TaskID, "", "", "", false); err != nil {
		t.Fatalf("RunTaskAssign: %v", err)
	}
	workers := internal.ListWorkers(dir, app.WtBase)
	if len(workers) != 1 {
		t.Fatalf("workers = %#v", workers)
	}
	cfg, _, err := internal.LoadWorkerConfigByID(dir, app.WtBase, workers[0].WorkerID)
	if err != nil {
		t.Fatalf("LoadWorkerConfigByID: %v", err)
	}
	if cfg.TaskID != record.TaskID {
		t.Fatalf("TaskID = %q, want %q", cfg.TaskID, record.TaskID)
	}
	if len(mock.SentTexts) < 2 {
		t.Fatalf("SentTexts = %#v, want launch + reminder", mock.SentTexts)
	}
	reminder := mock.SentTexts[len(mock.SentTexts)-1]
	if !strings.Contains(reminder, "Read worker.yaml first") {
		t.Fatalf("reminder = %q", reminder)
	}
	if strings.Contains(reminder, record.Title) {
		t.Fatalf("reminder should not contain task body/title: %q", reminder)
	}
}

func TestRunTaskAssignRebindRequiresSameRole(t *testing.T) {
	app, dir := initTestApp(t)
	record, err := internal.CreateTaskPackage(dir, "Implement feature", "backend", "", time.Now().UTC())
	if err != nil {
		t.Fatalf("CreateTaskPackage: %v", err)
	}
	worktreeCreated := false
	cfg := &internal.WorkerConfig{WorkerID: "frontend-001", Role: "frontend", Provider: "claude", WorktreeCreated: &worktreeCreated}
	if err := cfg.Save(internal.WorkerConfigPath(dir, "frontend-001")); err != nil {
		t.Fatalf("save worker config: %v", err)
	}
	err = app.RunTaskAssign(record.TaskID, "frontend-001", "", "", false)
	if err == nil || !strings.Contains(err.Error(), "role mismatch") {
		t.Fatalf("err = %v, want role mismatch", err)
	}
}

func TestRunWorkerAssignDelegatesToTaskAssign(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &MockBackend{SpawnedID: "pane-1", AlivePanes: map[string]bool{}}
	app.Session = mock
	roleDir := filepath.Join(dir, ".agents", "teams", "backend")
	refDir := filepath.Join(roleDir, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# backend\n"), 0644)
	os.WriteFile(filepath.Join(roleDir, "system.md"), []byte("# backend\n"), 0644)
	os.WriteFile(filepath.Join(refDir, "role.yaml"), []byte("name: backend\n"), 0644)
	skillInstaller = func(_, _, _, _, _ string, _ bool) error { return nil }
	taskSetup = func(string) error { return nil }
	workerShellInitDelay = 0
	defer func() {
		skillInstaller = internal.InstallSkillsForWorkerFromPath
		taskSetup = defaultTaskSetup
		workerShellInitDelay = 2 * time.Second
	}()
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
