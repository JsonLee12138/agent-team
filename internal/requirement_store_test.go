// internal/requirement_store_test.go
package internal

import (
	"os"
	"testing"
)

func TestRequirementPaths(t *testing.T) {
	wt := "/tmp/wt"
	if got := RequirementsDir(wt); got != "/tmp/wt/.tasks/requirements" {
		t.Errorf("RequirementsDir() = %s", got)
	}
	if got := RequirementDir(wt, "feat-auth"); got != "/tmp/wt/.tasks/requirements/feat-auth" {
		t.Errorf("RequirementDir() = %s", got)
	}
	if got := RequirementYAMLPath(wt, "feat-auth"); got != "/tmp/wt/.tasks/requirements/feat-auth/requirement.yaml" {
		t.Errorf("RequirementYAMLPath() = %s", got)
	}
	if got := RequirementIndexPath(wt); got != "/tmp/wt/.tasks/requirements/index.yaml" {
		t.Errorf("RequirementIndexPath() = %s", got)
	}
}

func TestSaveAndLoadRequirement(t *testing.T) {
	wtPath := t.TempDir()

	req := &Requirement{
		Name:        "feat-auth",
		Description: "Implement authentication",
		Status:      RequirementStatusOpen,
		CreatedAt:   "2026-03-18T00:00:00Z",
		SubTasks: []SubTask{
			{ID: 1, Title: "Design schema", Status: SubTaskStatusPending},
			{ID: 2, Title: "Implement API", Status: SubTaskStatusPending},
		},
	}

	// Save
	if err := SaveRequirement(wtPath, req); err != nil {
		t.Fatalf("SaveRequirement failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(RequirementYAMLPath(wtPath, "feat-auth")); os.IsNotExist(err) {
		t.Fatal("requirement.yaml not created")
	}

	// Load
	loaded, err := LoadRequirement(wtPath, "feat-auth")
	if err != nil {
		t.Fatalf("LoadRequirement failed: %v", err)
	}

	if loaded.Name != req.Name {
		t.Errorf("Name = %s, want %s", loaded.Name, req.Name)
	}
	if loaded.Status != RequirementStatusOpen {
		t.Errorf("Status = %s, want open", loaded.Status)
	}
	if len(loaded.SubTasks) != 2 {
		t.Errorf("SubTasks count = %d, want 2", len(loaded.SubTasks))
	}
	if loaded.SubTasks[0].Title != "Design schema" {
		t.Errorf("SubTask[0].Title = %s", loaded.SubTasks[0].Title)
	}
}

func TestLoadRequirementNotFound(t *testing.T) {
	wtPath := t.TempDir()

	_, err := LoadRequirement(wtPath, "nonexistent")
	if err == nil {
		t.Error("LoadRequirement should fail for nonexistent requirement")
	}
}

func TestListRequirements(t *testing.T) {
	wtPath := t.TempDir()

	// Empty list
	reqs, err := ListRequirements(wtPath)
	if err != nil {
		t.Fatalf("ListRequirements failed: %v", err)
	}
	if len(reqs) != 0 {
		t.Errorf("Expected 0 requirements, got %d", len(reqs))
	}

	// Create requirements
	names := []string{"c-third", "a-first", "b-second"}
	for _, name := range names {
		req := &Requirement{
			Name:      name,
			Status:    RequirementStatusOpen,
			CreatedAt: "2026-03-18T00:00:00Z",
		}
		if err := SaveRequirement(wtPath, req); err != nil {
			t.Fatalf("SaveRequirement failed: %v", err)
		}
	}

	reqs, err = ListRequirements(wtPath)
	if err != nil {
		t.Fatalf("ListRequirements failed: %v", err)
	}
	if len(reqs) != 3 {
		t.Fatalf("Expected 3 requirements, got %d", len(reqs))
	}

	// Verify sorted by name
	if reqs[0].Name != "a-first" {
		t.Errorf("First = %s, want a-first", reqs[0].Name)
	}
	if reqs[1].Name != "b-second" {
		t.Errorf("Second = %s, want b-second", reqs[1].Name)
	}
	if reqs[2].Name != "c-third" {
		t.Errorf("Third = %s, want c-third", reqs[2].Name)
	}
}

func TestCreateRequirement(t *testing.T) {
	wtPath := t.TempDir()

	subTasks := []SubTask{
		{ID: 1, Title: "Task A", Status: SubTaskStatusPending},
	}

	req, err := CreateRequirement(wtPath, "my-req", "My requirement", subTasks)
	if err != nil {
		t.Fatalf("CreateRequirement failed: %v", err)
	}

	if req.Status != RequirementStatusOpen {
		t.Errorf("Status = %s, want open", req.Status)
	}
	if req.CreatedAt == "" {
		t.Error("CreatedAt should be set")
	}

	// Load back
	loaded, err := LoadRequirement(wtPath, "my-req")
	if err != nil {
		t.Fatalf("LoadRequirement failed: %v", err)
	}
	if loaded.Description != "My requirement" {
		t.Errorf("Description = %s", loaded.Description)
	}
}

func TestRequirementIndexCRUD(t *testing.T) {
	wtPath := t.TempDir()

	// Load from non-existent returns empty
	idx, err := LoadRequirementIndex(wtPath)
	if err != nil {
		t.Fatalf("LoadRequirementIndex failed: %v", err)
	}
	if len(idx.Requirements) != 0 {
		t.Errorf("Expected empty index, got %d entries", len(idx.Requirements))
	}

	// Save and load
	idx.Requirements = []RequirementIndexEntry{
		{Name: "req-1", Status: RequirementStatusOpen, SubTaskCount: 3, DoneCount: 0},
	}
	if err := SaveRequirementIndex(wtPath, idx); err != nil {
		t.Fatalf("SaveRequirementIndex failed: %v", err)
	}

	loaded, err := LoadRequirementIndex(wtPath)
	if err != nil {
		t.Fatalf("LoadRequirementIndex failed: %v", err)
	}
	if len(loaded.Requirements) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(loaded.Requirements))
	}
	if loaded.Requirements[0].Name != "req-1" {
		t.Errorf("Name = %s", loaded.Requirements[0].Name)
	}
}

func TestUpdateIndexEntry(t *testing.T) {
	wtPath := t.TempDir()

	// Create a requirement
	req := &Requirement{
		Name:      "feat-x",
		Status:    RequirementStatusOpen,
		CreatedAt: "2026-03-18T00:00:00Z",
		SubTasks: []SubTask{
			{ID: 1, Title: "A", Status: SubTaskStatusPending},
			{ID: 2, Title: "B", Status: SubTaskStatusDone},
		},
	}
	if err := SaveRequirement(wtPath, req); err != nil {
		t.Fatalf("SaveRequirement failed: %v", err)
	}

	// Update index
	if err := UpdateIndexEntry(wtPath, req); err != nil {
		t.Fatalf("UpdateIndexEntry failed: %v", err)
	}

	idx, err := LoadRequirementIndex(wtPath)
	if err != nil {
		t.Fatalf("LoadRequirementIndex failed: %v", err)
	}
	if len(idx.Requirements) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(idx.Requirements))
	}
	if idx.Requirements[0].SubTaskCount != 2 {
		t.Errorf("SubTaskCount = %d, want 2", idx.Requirements[0].SubTaskCount)
	}
	if idx.Requirements[0].DoneCount != 1 {
		t.Errorf("DoneCount = %d, want 1", idx.Requirements[0].DoneCount)
	}

	// Update again (should replace, not append)
	req.SubTasks[0].Status = SubTaskStatusDone
	req.Status = RequirementStatusDone
	if err := UpdateIndexEntry(wtPath, req); err != nil {
		t.Fatalf("UpdateIndexEntry second call failed: %v", err)
	}

	idx, err = LoadRequirementIndex(wtPath)
	if err != nil {
		t.Fatalf("LoadRequirementIndex failed: %v", err)
	}
	if len(idx.Requirements) != 1 {
		t.Fatalf("Expected 1 entry after update, got %d", len(idx.Requirements))
	}
	if idx.Requirements[0].DoneCount != 2 {
		t.Errorf("DoneCount = %d, want 2", idx.Requirements[0].DoneCount)
	}
	if idx.Requirements[0].Status != RequirementStatusDone {
		t.Errorf("Status = %s, want done", idx.Requirements[0].Status)
	}
}

func TestRebuildRequirementIndex(t *testing.T) {
	wtPath := t.TempDir()

	// Create multiple requirements
	reqs := []*Requirement{
		{Name: "req-a", Status: RequirementStatusOpen, CreatedAt: "2026-03-18T00:00:00Z",
			SubTasks: []SubTask{{ID: 1, Title: "T1", Status: SubTaskStatusPending}}},
		{Name: "req-b", Status: RequirementStatusDone, CreatedAt: "2026-03-18T00:00:00Z",
			SubTasks: []SubTask{
				{ID: 1, Title: "T1", Status: SubTaskStatusDone},
				{ID: 2, Title: "T2", Status: SubTaskStatusSkipped},
			}},
	}

	for _, req := range reqs {
		if err := SaveRequirement(wtPath, req); err != nil {
			t.Fatalf("SaveRequirement failed: %v", err)
		}
	}

	idx, err := RebuildRequirementIndex(wtPath)
	if err != nil {
		t.Fatalf("RebuildRequirementIndex failed: %v", err)
	}
	if len(idx.Requirements) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(idx.Requirements))
	}

	// Verify it was saved
	loaded, err := LoadRequirementIndex(wtPath)
	if err != nil {
		t.Fatalf("LoadRequirementIndex failed: %v", err)
	}
	if len(loaded.Requirements) != 2 {
		t.Fatalf("Saved index has %d entries, want 2", len(loaded.Requirements))
	}

	// req-a: 1 subtask, 0 done
	// req-b: 2 subtasks, 2 done (done + skipped both count)
	for _, e := range loaded.Requirements {
		switch e.Name {
		case "req-a":
			if e.SubTaskCount != 1 || e.DoneCount != 0 {
				t.Errorf("req-a: SubTaskCount=%d DoneCount=%d", e.SubTaskCount, e.DoneCount)
			}
		case "req-b":
			if e.SubTaskCount != 2 || e.DoneCount != 2 {
				t.Errorf("req-b: SubTaskCount=%d DoneCount=%d", e.SubTaskCount, e.DoneCount)
			}
		}
	}
}
