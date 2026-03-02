// internal/task_store_test.go
package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitTasksDir(t *testing.T) {
	wtPath := t.TempDir()

	// First call should create everything
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	// Check that directories were created
	if _, err := os.Stat(TasksDir(wtPath)); os.IsNotExist(err) {
		t.Error("tasks directory not created")
	}
	if _, err := os.Stat(TasksChangesDir(wtPath)); os.IsNotExist(err) {
		t.Error("changes directory not created")
	}

	// Check that config was created
	if _, err := os.Stat(TasksConfigPath(wtPath)); os.IsNotExist(err) {
		t.Error("config.yaml not created")
	}

	// Second call should be idempotent
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("Second InitTasksDir failed: %v", err)
	}
}

func TestCreateTaskChange(t *testing.T) {
	wtPath := t.TempDir()
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	changeName := "2024-01-01-test-feature"
	description := "Test Feature"
	proposal := "# Proposal\n\nTest proposal content"
	design := "# Design\n\nTest design content"

	// Create change
	changeDir, err := CreateTaskChange(wtPath, changeName, description, proposal, design)
	if err != nil {
		t.Fatalf("CreateTaskChange failed: %v", err)
	}

	// Check that change directory exists
	if _, err := os.Stat(changeDir); os.IsNotExist(err) {
		t.Error("change directory not created")
	}

	// Check that change.yaml exists
	if _, err := os.Stat(ChangeYAMLPath(wtPath, changeName)); os.IsNotExist(err) {
		t.Error("change.yaml not created")
	}

	// Check that proposal.md exists
	if _, err := os.Stat(filepath.Join(changeDir, "proposal.md")); os.IsNotExist(err) {
		t.Error("proposal.md not created")
	}

	// Check that design.md exists
	if _, err := os.Stat(filepath.Join(changeDir, "design.md")); os.IsNotExist(err) {
		t.Error("design.md not created")
	}

	// Load and verify change
	change, err := LoadChange(wtPath, changeName)
	if err != nil {
		t.Fatalf("LoadChange failed: %v", err)
	}

	if change.Name != changeName {
		t.Errorf("Name mismatch: %s != %s", change.Name, changeName)
	}
	if change.Status != ChangeStatusDraft {
		t.Errorf("Status should be draft, got %s", change.Status)
	}
	if change.Description != description {
		t.Errorf("Description mismatch: %s != %s", change.Description, description)
	}
}

func TestCreateTaskChangeWithoutProposalOrDesign(t *testing.T) {
	wtPath := t.TempDir()
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	changeName := "2024-01-01-minimal-change"
	description := "Minimal Change"

	// Create change without proposal or design
	changeDir, err := CreateTaskChange(wtPath, changeName, description, "", "")
	if err != nil {
		t.Fatalf("CreateTaskChange failed: %v", err)
	}

	// Check that proposal.md does NOT exist
	if _, err := os.Stat(filepath.Join(changeDir, "proposal.md")); err == nil {
		t.Error("proposal.md should not be created when empty")
	}

	// Check that design.md does NOT exist
	if _, err := os.Stat(filepath.Join(changeDir, "design.md")); err == nil {
		t.Error("design.md should not be created when empty")
	}
}

func TestLoadAndSaveChange(t *testing.T) {
	wtPath := t.TempDir()
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	// Create and load a change
	changeName := "2024-01-01-test"
	_, err := CreateTaskChange(wtPath, changeName, "Test", "", "")
	if err != nil {
		t.Fatalf("CreateTaskChange failed: %v", err)
	}

	// Modify the change
	change, err := LoadChange(wtPath, changeName)
	if err != nil {
		t.Fatalf("LoadChange failed: %v", err)
	}

	change.Status = ChangeStatusAssigned
	change.AssignedTo = "worker1"
	change.Tasks = []Task{
		{ID: 1, Title: "Task 1", Status: TaskStatusPending},
	}

	// Save the modified change
	if err := SaveChange(wtPath, change); err != nil {
		t.Fatalf("SaveChange failed: %v", err)
	}

	// Load again and verify
	reloaded, err := LoadChange(wtPath, changeName)
	if err != nil {
		t.Fatalf("LoadChange after save failed: %v", err)
	}

	if reloaded.Status != ChangeStatusAssigned {
		t.Errorf("Status not saved: got %s", reloaded.Status)
	}
	if reloaded.AssignedTo != "worker1" {
		t.Errorf("AssignedTo not saved: got %s", reloaded.AssignedTo)
	}
	if len(reloaded.Tasks) != 1 {
		t.Errorf("Tasks not saved: got %d tasks", len(reloaded.Tasks))
	}
}

func TestListChanges(t *testing.T) {
	wtPath := t.TempDir()
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	// Create multiple changes
	names := []string{
		"2024-01-03-third",
		"2024-01-01-first",
		"2024-01-02-second",
	}

	for _, name := range names {
		_, err := CreateTaskChange(wtPath, name, "Test "+name, "", "")
		if err != nil {
			t.Fatalf("CreateTaskChange failed: %v", err)
		}
	}

	// List all changes
	changes, err := ListChanges(wtPath)
	if err != nil {
		t.Fatalf("ListChanges failed: %v", err)
	}

	if len(changes) != 3 {
		t.Errorf("Expected 3 changes, got %d", len(changes))
	}

	// Verify they are sorted
	if changes[0].Name != "2024-01-01-first" {
		t.Errorf("First change should be 2024-01-01-first, got %s", changes[0].Name)
	}
	if changes[1].Name != "2024-01-02-second" {
		t.Errorf("Second change should be 2024-01-02-second, got %s", changes[1].Name)
	}
	if changes[2].Name != "2024-01-03-third" {
		t.Errorf("Third change should be 2024-01-03-third, got %s", changes[2].Name)
	}
}

func TestListActiveChanges(t *testing.T) {
	wtPath := t.TempDir()
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	// Create changes with different statuses
	changes := []*Change{
		{Name: "change1", Description: "Test", Status: ChangeStatusDraft, CreatedAt: "2024-01-01T00:00:00Z"},
		{Name: "change2", Description: "Test", Status: ChangeStatusArchived, CreatedAt: "2024-01-01T00:00:00Z"},
		{Name: "change3", Description: "Test", Status: ChangeStatusDone, CreatedAt: "2024-01-01T00:00:00Z"},
	}

	for _, c := range changes {
		if err := SaveChange(wtPath, c); err != nil {
			t.Fatalf("SaveChange failed: %v", err)
		}
		// Create the directory
		os.MkdirAll(ChangeDirPath(wtPath, c.Name), 0755)
	}

	// List active changes
	active, err := ListActiveChanges(wtPath)
	if err != nil {
		t.Fatalf("ListActiveChanges failed: %v", err)
	}

	if len(active) != 2 {
		t.Errorf("Expected 2 active changes, got %d", len(active))
	}

	// Verify archived is not included
	for _, c := range active {
		if c.Status == ChangeStatusArchived {
			t.Error("Archived change should not be in active list")
		}
	}
}

func TestCountActiveChanges(t *testing.T) {
	wtPath := t.TempDir()
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	// Create some changes
	for i := 0; i < 3; i++ {
		name := "change" + string(rune('1'+i))
		_, err := CreateTaskChange(wtPath, name, "Test", "", "")
		if err != nil {
			t.Fatalf("CreateTaskChange failed: %v", err)
		}
	}

	count := CountActiveChanges(wtPath)
	if count != 3 {
		t.Errorf("Expected 3 active changes, got %d", count)
	}
}

func TestListChangesEmptyDirectory(t *testing.T) {
	wtPath := t.TempDir()
	if err := InitTasksDir(wtPath); err != nil {
		t.Fatalf("InitTasksDir failed: %v", err)
	}

	changes, err := ListChanges(wtPath)
	if err != nil {
		t.Fatalf("ListChanges failed: %v", err)
	}

	if len(changes) != 0 {
		t.Errorf("Expected 0 changes, got %d", len(changes))
	}
}

func TestListChangesNonExistentTasksDir(t *testing.T) {
	wtPath := t.TempDir()
	// Don't call InitTasksDir, so .tasks doesn't exist

	changes, err := ListChanges(wtPath)
	if err != nil {
		t.Fatalf("ListChanges should not error for non-existent directory: %v", err)
	}

	if len(changes) != 0 {
		t.Errorf("Expected 0 changes, got %d", len(changes))
	}
}
