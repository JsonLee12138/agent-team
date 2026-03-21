package cmd

import (
	"os"
	"path/filepath"
	"strings"
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
	verificationPath := internal.TaskVerificationPath(dir, archived.TaskID)
	if err := os.WriteFile(verificationPath, []byte("# Verification\n\n## Result\n- pass\n"), 0644); err != nil {
		t.Fatalf("WriteFile verification: %v", err)
	}
	if _, err := internal.ArchiveTask(dir, archived.TaskID, "abc123", false, now.Add(4*time.Minute)); err != nil {
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

func TestRunTaskListPrintsVerificationColumns(t *testing.T) {
	app, dir := initTestApp(t)
	record, err := internal.CreateTaskPackage(dir, "Show Columns", "backend", "", time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("CreateTaskPackage: %v", err)
	}
	if _, err := internal.BindTaskToWorker(dir, record.TaskID, "backend-001", time.Now().UTC()); err != nil {
		t.Fatalf("BindTaskToWorker: %v", err)
	}
	if _, err := internal.MarkTaskDone(dir, record.TaskID, time.Now().UTC()); err != nil {
		t.Fatalf("MarkTaskDone: %v", err)
	}
	if err := os.WriteFile(internal.TaskVerificationPath(dir, record.TaskID), []byte("# Verification\n\n## Result\n- partial\n"), 0644); err != nil {
		t.Fatalf("WriteFile verification: %v", err)
	}
	out := captureStdout(t, func() {
		if err := app.RunTaskList(false); err != nil {
			t.Fatalf("RunTaskList: %v", err)
		}
	})
	for _, needle := range []string{"Verification", "Archive Ready", "partial", "strict-no"} {
		if !strings.Contains(out, needle) {
			t.Fatalf("output missing %q:\n%s", needle, out)
		}
	}
}
