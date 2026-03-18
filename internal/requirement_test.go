// internal/requirement_test.go
package internal

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestValidRequirementStatus(t *testing.T) {
	tests := []struct {
		status RequirementStatus
		valid  bool
	}{
		{RequirementStatusOpen, true},
		{RequirementStatusInProgress, true},
		{RequirementStatusDone, true},
		{RequirementStatus("invalid"), false},
		{RequirementStatus(""), false},
	}

	for _, tt := range tests {
		if got := ValidRequirementStatus(tt.status); got != tt.valid {
			t.Errorf("ValidRequirementStatus(%q) = %v, want %v", tt.status, got, tt.valid)
		}
	}
}

func TestValidSubTaskStatus(t *testing.T) {
	tests := []struct {
		status SubTaskStatus
		valid  bool
	}{
		{SubTaskStatusPending, true},
		{SubTaskStatusAssigned, true},
		{SubTaskStatusDone, true},
		{SubTaskStatusSkipped, true},
		{SubTaskStatus("invalid"), false},
		{SubTaskStatus(""), false},
	}

	for _, tt := range tests {
		if got := ValidSubTaskStatus(tt.status); got != tt.valid {
			t.Errorf("ValidSubTaskStatus(%q) = %v, want %v", tt.status, got, tt.valid)
		}
	}
}

func TestRequirementYAMLRoundtrip(t *testing.T) {
	req := Requirement{
		Name:        "feat-auth",
		Description: "Implement user authentication",
		Status:      RequirementStatusInProgress,
		CreatedAt:   "2026-03-18T10:00:00Z",
		SubTasks: []SubTask{
			{ID: 1, Title: "Design schema", Status: SubTaskStatusDone},
			{ID: 2, Title: "Implement API", Status: SubTaskStatusAssigned, AssignedTo: "backend-001", ChangeName: "auth-api"},
			{ID: 3, Title: "Write tests", Status: SubTaskStatusPending},
		},
	}

	data, err := yaml.Marshal(&req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var loaded Requirement
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if loaded.Name != req.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, req.Name)
	}
	if loaded.Description != req.Description {
		t.Errorf("Description = %q, want %q", loaded.Description, req.Description)
	}
	if loaded.Status != req.Status {
		t.Errorf("Status = %q, want %q", loaded.Status, req.Status)
	}
	if loaded.CreatedAt != req.CreatedAt {
		t.Errorf("CreatedAt = %q, want %q", loaded.CreatedAt, req.CreatedAt)
	}
	if len(loaded.SubTasks) != 3 {
		t.Fatalf("SubTasks count = %d, want 3", len(loaded.SubTasks))
	}
	if loaded.SubTasks[1].AssignedTo != "backend-001" {
		t.Errorf("SubTask[1].AssignedTo = %q, want backend-001", loaded.SubTasks[1].AssignedTo)
	}
	if loaded.SubTasks[1].ChangeName != "auth-api" {
		t.Errorf("SubTask[1].ChangeName = %q, want auth-api", loaded.SubTasks[1].ChangeName)
	}
}

func TestRequirementYAMLOmitEmpty(t *testing.T) {
	// SubTasks with omitempty: empty slice should not appear in YAML
	req := Requirement{
		Name:   "minimal",
		Status: RequirementStatusOpen,
	}

	data, err := yaml.Marshal(&req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	content := string(data)
	// sub_tasks should be omitted when nil
	if got := string(data); got != content {
		t.Errorf("unexpected content")
	}

	var loaded Requirement
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(loaded.SubTasks) != 0 {
		t.Errorf("SubTasks should be empty, got %d", len(loaded.SubTasks))
	}
}

func TestSubTaskYAMLOmitEmptyFields(t *testing.T) {
	// AssignedTo and ChangeName are omitempty — should not appear in YAML when empty
	st := SubTask{ID: 1, Title: "Task", Status: SubTaskStatusPending}

	data, err := yaml.Marshal(&st)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	content := string(data)
	for _, field := range []string{"assigned_to", "change_name"} {
		if strings.Contains(content, field) {
			t.Errorf("YAML should omit %q when empty, got:\n%s", field, content)
		}
	}
}

func TestRequirementYAMLFromRawString(t *testing.T) {
	// Simulate loading from a hand-written YAML file
	raw := `name: my-feature
description: "Build the feature"
status: open
created_at: "2026-01-01T00:00:00Z"
sub_tasks:
  - id: 1
    title: "First task"
    status: pending
  - id: 2
    title: "Second task"
    assigned_to: worker-x
    status: assigned
    change_name: impl-second
`

	var req Requirement
	if err := yaml.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("Unmarshal raw YAML failed: %v", err)
	}
	if req.Name != "my-feature" {
		t.Errorf("Name = %q", req.Name)
	}
	if req.Status != RequirementStatusOpen {
		t.Errorf("Status = %q", req.Status)
	}
	if len(req.SubTasks) != 2 {
		t.Fatalf("SubTasks count = %d, want 2", len(req.SubTasks))
	}
	if req.SubTasks[1].AssignedTo != "worker-x" {
		t.Errorf("SubTask[1].AssignedTo = %q", req.SubTasks[1].AssignedTo)
	}
	if req.SubTasks[1].ChangeName != "impl-second" {
		t.Errorf("SubTask[1].ChangeName = %q", req.SubTasks[1].ChangeName)
	}
}

func TestRequirementIndexYAMLRoundtrip(t *testing.T) {
	idx := RequirementIndex{
		Requirements: []RequirementIndexEntry{
			{Name: "req-a", Status: RequirementStatusOpen, SubTaskCount: 3, DoneCount: 0},
			{Name: "req-b", Status: RequirementStatusDone, SubTaskCount: 2, DoneCount: 2},
		},
	}

	data, err := yaml.Marshal(&idx)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var loaded RequirementIndex
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(loaded.Requirements) != 2 {
		t.Fatalf("Requirements count = %d, want 2", len(loaded.Requirements))
	}
	if loaded.Requirements[0].Name != "req-a" {
		t.Errorf("[0].Name = %q", loaded.Requirements[0].Name)
	}
	if loaded.Requirements[1].DoneCount != 2 {
		t.Errorf("[1].DoneCount = %d, want 2", loaded.Requirements[1].DoneCount)
	}
}
