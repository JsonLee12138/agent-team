// internal/requirement_lifecycle_test.go
package internal

import (
	"testing"
)

func TestValidateSubTaskTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    SubTaskStatus
		to      SubTaskStatus
		wantErr bool
	}{
		// Valid transitions
		{name: "pending → assigned", from: SubTaskStatusPending, to: SubTaskStatusAssigned, wantErr: false},
		{name: "pending → skipped", from: SubTaskStatusPending, to: SubTaskStatusSkipped, wantErr: false},
		{name: "assigned → done", from: SubTaskStatusAssigned, to: SubTaskStatusDone, wantErr: false},
		{name: "assigned → skipped", from: SubTaskStatusAssigned, to: SubTaskStatusSkipped, wantErr: false},
		{name: "assigned → pending", from: SubTaskStatusAssigned, to: SubTaskStatusPending, wantErr: false},

		// Invalid transitions
		{name: "pending → done", from: SubTaskStatusPending, to: SubTaskStatusDone, wantErr: true},
		{name: "done → pending", from: SubTaskStatusDone, to: SubTaskStatusPending, wantErr: true},
		{name: "done → assigned", from: SubTaskStatusDone, to: SubTaskStatusAssigned, wantErr: true},
		{name: "skipped → pending", from: SubTaskStatusSkipped, to: SubTaskStatusPending, wantErr: true},

		// Unknown status
		{name: "unknown → pending", from: SubTaskStatus("unknown"), to: SubTaskStatusPending, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSubTaskTransition(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSubTaskTransition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAutoPromoteRequirement(t *testing.T) {
	tests := []struct {
		name     string
		req      *Requirement
		promoted bool
		wantSt   RequirementStatus
	}{
		{
			name:     "no sub-tasks → no promote",
			req:      &Requirement{Status: RequirementStatusOpen},
			promoted: false,
			wantSt:   RequirementStatusOpen,
		},
		{
			name: "all done → promote",
			req: &Requirement{
				Status: RequirementStatusInProgress,
				SubTasks: []SubTask{
					{ID: 1, Status: SubTaskStatusDone},
					{ID: 2, Status: SubTaskStatusDone},
				},
			},
			promoted: true,
			wantSt:   RequirementStatusDone,
		},
		{
			name: "all skipped → promote",
			req: &Requirement{
				Status: RequirementStatusInProgress,
				SubTasks: []SubTask{
					{ID: 1, Status: SubTaskStatusSkipped},
				},
			},
			promoted: true,
			wantSt:   RequirementStatusDone,
		},
		{
			name: "mixed done and skipped → promote",
			req: &Requirement{
				Status: RequirementStatusInProgress,
				SubTasks: []SubTask{
					{ID: 1, Status: SubTaskStatusDone},
					{ID: 2, Status: SubTaskStatusSkipped},
				},
			},
			promoted: true,
			wantSt:   RequirementStatusDone,
		},
		{
			name: "has pending → no promote",
			req: &Requirement{
				Status: RequirementStatusInProgress,
				SubTasks: []SubTask{
					{ID: 1, Status: SubTaskStatusDone},
					{ID: 2, Status: SubTaskStatusPending},
				},
			},
			promoted: false,
			wantSt:   RequirementStatusInProgress,
		},
		{
			name: "already done → no promote",
			req: &Requirement{
				Status: RequirementStatusDone,
				SubTasks: []SubTask{
					{ID: 1, Status: SubTaskStatusDone},
				},
			},
			promoted: false,
			wantSt:   RequirementStatusDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AutoPromoteRequirement(tt.req)
			if got != tt.promoted {
				t.Errorf("AutoPromoteRequirement() = %v, want %v", got, tt.promoted)
			}
			if tt.req.Status != tt.wantSt {
				t.Errorf("Status = %s, want %s", tt.req.Status, tt.wantSt)
			}
		})
	}
}

func TestMarkSubTaskDone(t *testing.T) {
	t.Run("marks sub-task done and auto-promotes", func(t *testing.T) {
		req := &Requirement{
			Status: RequirementStatusInProgress,
			SubTasks: []SubTask{
				{ID: 1, Status: SubTaskStatusAssigned},
				{ID: 2, Status: SubTaskStatusDone},
			},
		}

		if err := MarkSubTaskDone(req, 1); err != nil {
			t.Fatalf("MarkSubTaskDone failed: %v", err)
		}

		if req.SubTasks[0].Status != SubTaskStatusDone {
			t.Errorf("SubTask[0].Status = %s, want done", req.SubTasks[0].Status)
		}
		if req.Status != RequirementStatusDone {
			t.Errorf("Requirement status = %s, want done (auto-promote)", req.Status)
		}
	})

	t.Run("marks sub-task done without promote", func(t *testing.T) {
		req := &Requirement{
			Status: RequirementStatusInProgress,
			SubTasks: []SubTask{
				{ID: 1, Status: SubTaskStatusAssigned},
				{ID: 2, Status: SubTaskStatusPending},
			},
		}

		if err := MarkSubTaskDone(req, 1); err != nil {
			t.Fatalf("MarkSubTaskDone failed: %v", err)
		}

		if req.SubTasks[0].Status != SubTaskStatusDone {
			t.Errorf("SubTask[0].Status = %s, want done", req.SubTasks[0].Status)
		}
		if req.Status != RequirementStatusInProgress {
			t.Errorf("Requirement status = %s, want in_progress", req.Status)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := &Requirement{SubTasks: []SubTask{{ID: 1, Status: SubTaskStatusAssigned}}}
		if err := MarkSubTaskDone(req, 99); err == nil {
			t.Error("Expected error for nonexistent sub-task")
		}
	})

	t.Run("invalid transition", func(t *testing.T) {
		req := &Requirement{SubTasks: []SubTask{{ID: 1, Status: SubTaskStatusPending}}}
		if err := MarkSubTaskDone(req, 1); err == nil {
			t.Error("Expected error for pending → done transition")
		}
	})
}

func TestAssignSubTask(t *testing.T) {
	t.Run("assigns sub-task and promotes requirement to in_progress", func(t *testing.T) {
		req := &Requirement{
			Status: RequirementStatusOpen,
			SubTasks: []SubTask{
				{ID: 1, Title: "Task A", Status: SubTaskStatusPending},
				{ID: 2, Title: "Task B", Status: SubTaskStatusPending},
			},
		}

		if err := AssignSubTask(req, 1, "worker-1", "change-abc"); err != nil {
			t.Fatalf("AssignSubTask failed: %v", err)
		}

		if req.SubTasks[0].Status != SubTaskStatusAssigned {
			t.Errorf("SubTask.Status = %s, want assigned", req.SubTasks[0].Status)
		}
		if req.SubTasks[0].AssignedTo != "worker-1" {
			t.Errorf("AssignedTo = %s, want worker-1", req.SubTasks[0].AssignedTo)
		}
		if req.SubTasks[0].ChangeName != "change-abc" {
			t.Errorf("ChangeName = %s, want change-abc", req.SubTasks[0].ChangeName)
		}
		if req.Status != RequirementStatusInProgress {
			t.Errorf("Requirement.Status = %s, want in_progress", req.Status)
		}
	})

	t.Run("keeps in_progress if already in_progress", func(t *testing.T) {
		req := &Requirement{
			Status: RequirementStatusInProgress,
			SubTasks: []SubTask{
				{ID: 1, Status: SubTaskStatusPending},
			},
		}

		if err := AssignSubTask(req, 1, "w2", "c2"); err != nil {
			t.Fatalf("AssignSubTask failed: %v", err)
		}
		if req.Status != RequirementStatusInProgress {
			t.Errorf("Status = %s, want in_progress", req.Status)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := &Requirement{SubTasks: []SubTask{{ID: 1, Status: SubTaskStatusPending}}}
		if err := AssignSubTask(req, 99, "w", "c"); err == nil {
			t.Error("Expected error for nonexistent sub-task")
		}
	})

	t.Run("invalid transition", func(t *testing.T) {
		req := &Requirement{SubTasks: []SubTask{{ID: 1, Status: SubTaskStatusDone}}}
		if err := AssignSubTask(req, 1, "w", "c"); err == nil {
			t.Error("Expected error for done → assigned transition")
		}
	})
}
