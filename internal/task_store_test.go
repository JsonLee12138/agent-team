package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreateTaskPackage(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	record, err := CreateTaskPackage(root, "Implement task lifecycle", "backend", "## Design\n\n- approved", now)
	if err != nil {
		t.Fatalf("CreateTaskPackage: %v", err)
	}
	if record.Status != TaskStatusDraft {
		t.Fatalf("Status = %s", record.Status)
	}
	if !strings.HasPrefix(record.TaskID, "2026-03-21-10-00-00-") {
		t.Fatalf("TaskID = %s", record.TaskID)
	}
	data, err := os.ReadFile(TaskContextPath(root, record.TaskID))
	if err != nil {
		t.Fatalf("ReadFile context: %v", err)
	}
	if !strings.Contains(string(data), "## Design") {
		t.Fatalf("context missing design section: %s", string(data))
	}
}

func TestListTasksActiveOnly(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	active, _ := CreateTaskPackage(root, "Active Task", "backend", "", now)
	archived, _ := CreateTaskPackage(root, "Archived Task", "backend", "", now.Add(time.Minute))
	if _, err := BindTaskToWorker(root, archived.TaskID, "backend-001", now.Add(2*time.Minute)); err != nil {
		t.Fatalf("BindTaskToWorker archived: %v", err)
	}
	if _, err := MarkTaskDone(root, archived.TaskID, now.Add(3*time.Minute)); err != nil {
		t.Fatalf("MarkTaskDone archived: %v", err)
	}
	if _, err := ArchiveTask(root, archived.TaskID, "abc123", now.Add(4*time.Minute)); err != nil {
		t.Fatalf("ArchiveTask: %v", err)
	}
	list, err := ListTasks(root, true)
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(list) != 1 || list[0].TaskID != active.TaskID {
		t.Fatalf("active list = %#v", list)
	}
}

func TestBindTaskToWorkerAllowsReassignFromDone(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	record, err := CreateTaskPackage(root, "Rework Task", "backend", "", now)
	if err != nil {
		t.Fatalf("CreateTaskPackage: %v", err)
	}
	if _, err := BindTaskToWorker(root, record.TaskID, "backend-001", now.Add(time.Minute)); err != nil {
		t.Fatalf("BindTaskToWorker: %v", err)
	}
	if _, err := MarkTaskDone(root, record.TaskID, now.Add(2*time.Minute)); err != nil {
		t.Fatalf("MarkTaskDone: %v", err)
	}
	record, err = BindTaskToWorker(root, record.TaskID, "backend-002", now.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("BindTaskToWorker reassign: %v", err)
	}
	if record.WorkerID != "backend-002" || record.Status != TaskStatusAssigned {
		t.Fatalf("record after reassign = %#v", record)
	}
}

func TestArchiveTaskRollsBackOnWriteFailure(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	record, err := CreateTaskPackage(root, "Archive Rollback", "backend", "", now)
	if err != nil {
		t.Fatalf("CreateTaskPackage: %v", err)
	}
	if _, err := BindTaskToWorker(root, record.TaskID, "backend-001", now.Add(time.Minute)); err != nil {
		t.Fatalf("BindTaskToWorker: %v", err)
	}
	if _, err := MarkTaskDone(root, record.TaskID, now.Add(2*time.Minute)); err != nil {
		t.Fatalf("MarkTaskDone: %v", err)
	}

	orig := saveTaskRecordAt
	saveTaskRecordAt = func(dir string, record *TaskRecord) error {
		if strings.Contains(filepath.ToSlash(dir), filepath.ToSlash(TaskArchiveDir(root, record.TaskID))) {
			return fmt.Errorf("forced write failure")
		}
		return orig(dir, record)
	}
	defer func() { saveTaskRecordAt = orig }()

	_, err = ArchiveTask(root, record.TaskID, "deadbeef", now.Add(3*time.Minute))
	if err == nil {
		t.Fatal("expected archive failure")
	}
	if _, statErr := os.Stat(TaskDir(root, record.TaskID)); statErr != nil {
		t.Fatalf("task dir should be restored: %v", statErr)
	}
	if _, statErr := os.Stat(TaskArchiveDir(root, record.TaskID)); !os.IsNotExist(statErr) {
		t.Fatalf("archive dir should not remain after rollback: %v", statErr)
	}
	loaded, archived, loadErr := LoadTaskRecord(root, record.TaskID)
	if loadErr != nil {
		t.Fatalf("LoadTaskRecord: %v", loadErr)
	}
	if archived || loaded.Status != TaskStatusDone {
		t.Fatalf("loaded after rollback = %#v archived=%v", loaded, archived)
	}
}
