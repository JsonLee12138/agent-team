package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunTaskArchiveMovesTaskAndCleansWorker(t *testing.T) {
	app, dir := initTestApp(t)
	app.Session = &closeMockBackend{alivePanes: map[string]bool{}}
	workerID := "backend-001"
	wtPath := filepath.Join(dir, app.WtBase, workerID)
	if err := app.Git.WorktreeAdd(wtPath, "team/"+workerID); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}
	worktreeCreated := true
	cfg := &internal.WorkerConfig{WorkerID: workerID, Role: "backend", Provider: "claude", WorktreeCreated: &worktreeCreated}
	if err := cfg.Save(internal.WorkerYAMLPath(wtPath)); err != nil {
		t.Fatalf("save worker config: %v", err)
	}
	record, err := internal.CreateTaskPackage(dir, "Archive Task", "backend", "", time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("CreateTaskPackage: %v", err)
	}
	if _, err := internal.BindTaskToWorker(dir, record.TaskID, workerID, time.Now().UTC()); err != nil {
		t.Fatalf("BindTaskToWorker: %v", err)
	}
	if _, err := internal.MarkTaskDone(dir, record.TaskID, time.Now().UTC()); err != nil {
		t.Fatalf("MarkTaskDone: %v", err)
	}
	if err := app.RunTaskArchive(record.TaskID, "deadbeef"); err != nil {
		t.Fatalf("RunTaskArchive: %v", err)
	}
	if _, err := os.Stat(internal.TaskArchiveDir(dir, record.TaskID)); err != nil {
		t.Fatalf("archived task missing: %v", err)
	}
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Fatalf("worktree should be removed, err=%v", err)
	}
}
