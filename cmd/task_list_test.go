package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunTaskListShowsActiveByDefault(t *testing.T) {
	_, dir := initTestApp(t)
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	active, _ := internal.CreateTaskPackage(dir, "Active", "backend", "", now)
	archived, _ := internal.CreateTaskPackage(dir, "Archived", "backend", "", now.Add(time.Minute))
	if _, err := internal.BindTaskToWorker(dir, archived.TaskID, "backend-001", now.Add(2*time.Minute)); err != nil {
		t.Fatalf("BindTaskToWorker: %v", err)
	}
	if _, err := internal.MarkTaskDone(dir, archived.TaskID, now.Add(3*time.Minute)); err != nil {
		t.Fatalf("MarkTaskDone: %v", err)
	}
	if _, err := internal.ArchiveTask(dir, archived.TaskID, "abc123", now.Add(4*time.Minute)); err != nil {
		t.Fatalf("ArchiveTask: %v", err)
	}
	list, err := internal.ListTasks(dir, true)
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(list) != 1 || list[0].TaskID != active.TaskID {
		t.Fatalf("list = %#v", list)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agent-team", "archive", "task", archived.TaskID)); err != nil {
		t.Fatalf("archived task dir missing: %v", err)
	}
}
