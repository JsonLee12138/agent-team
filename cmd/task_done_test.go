package cmd

import (
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunTaskDoneMarksAssignedTaskVerifying(t *testing.T) {
	app, dir := initTestApp(t)
	record, err := internal.CreateTaskPackage(dir, "Done Task", "backend", "", time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("CreateTaskPackage: %v", err)
	}
	if _, err := internal.BindTaskToWorker(dir, record.TaskID, "backend-001", time.Now().UTC()); err != nil {
		t.Fatalf("BindTaskToWorker: %v", err)
	}
	if err := app.RunTaskDone(record.TaskID); err != nil {
		t.Fatalf("RunTaskDone: %v", err)
	}
	loaded, _, err := internal.LoadTaskRecord(dir, record.TaskID)
	if err != nil {
		t.Fatalf("LoadTaskRecord: %v", err)
	}
	if loaded.Status != internal.TaskStatusVerifying {
		t.Fatalf("Status = %s", loaded.Status)
	}
}
