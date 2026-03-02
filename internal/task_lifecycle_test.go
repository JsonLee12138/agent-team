// internal/task_lifecycle_test.go
package internal

import (
	"testing"
)

func TestValidateChangeTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    ChangeStatus
		to      ChangeStatus
		wantErr bool
	}{
		// Valid transitions
		{name: "draft → assigned", from: ChangeStatusDraft, to: ChangeStatusAssigned, wantErr: false},
		{name: "draft → archived", from: ChangeStatusDraft, to: ChangeStatusArchived, wantErr: false},
		{name: "assigned → implementing", from: ChangeStatusAssigned, to: ChangeStatusImplementing, wantErr: false},
		{name: "assigned → draft", from: ChangeStatusAssigned, to: ChangeStatusDraft, wantErr: false},
		{name: "implementing → verifying", from: ChangeStatusImplementing, to: ChangeStatusVerifying, wantErr: false},
		{name: "verifying → done", from: ChangeStatusVerifying, to: ChangeStatusDone, wantErr: false},
		{name: "verifying → implementing", from: ChangeStatusVerifying, to: ChangeStatusImplementing, wantErr: false},
		{name: "done → archived", from: ChangeStatusDone, to: ChangeStatusArchived, wantErr: false},

		// Invalid transitions
		{name: "draft → verifying", from: ChangeStatusDraft, to: ChangeStatusVerifying, wantErr: true},
		{name: "archived → draft", from: ChangeStatusArchived, to: ChangeStatusDraft, wantErr: true},
		{name: "done → implementing", from: ChangeStatusDone, to: ChangeStatusImplementing, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChangeTransition(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateChangeTransition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplyChangeTransition(t *testing.T) {
	change := &Change{
		Name:   "test-change",
		Status: ChangeStatusDraft,
	}

	// Valid transition
	if err := ApplyChangeTransition(change, ChangeStatusAssigned); err != nil {
		t.Errorf("ApplyChangeTransition failed: %v", err)
	}
	if change.Status != ChangeStatusAssigned {
		t.Errorf("Status not updated: %s", change.Status)
	}

	// Valid transition (assigned → draft is allowed)
	if err := ApplyChangeTransition(change, ChangeStatusDraft); err != nil {
		t.Errorf("ApplyChangeTransition should allow assigned → draft: %v", err)
	}
	if change.Status != ChangeStatusDraft {
		t.Errorf("Status not updated: %s", change.Status)
	}

	// Reset to assigned
	change.Status = ChangeStatusAssigned

	// Invalid transition (assigned → done is not allowed)
	if err := ApplyChangeTransition(change, ChangeStatusDone); err == nil {
		t.Error("ApplyChangeTransition should fail for invalid transition (assigned → done)")
	}
	if change.Status != ChangeStatusAssigned {
		t.Errorf("Status should not change after failed transition: %s", change.Status)
	}
}

func TestIsTerminalChangeStatus(t *testing.T) {
	tests := []struct {
		status   ChangeStatus
		terminal bool
	}{
		{ChangeStatusDraft, false},
		{ChangeStatusAssigned, false},
		{ChangeStatusImplementing, false},
		{ChangeStatusVerifying, false},
		{ChangeStatusDone, false},
		{ChangeStatusArchived, true},
	}

	for _, tt := range tests {
		if IsTerminalChangeStatus(tt.status) != tt.terminal {
			t.Errorf("IsTerminalChangeStatus(%s) = %v, want %v",
				tt.status, IsTerminalChangeStatus(tt.status), tt.terminal)
		}
	}
}

func TestAllTasksDone(t *testing.T) {
	tests := []struct {
		name  string
		tasks []Task
		want  bool
	}{
		{
			name:  "empty task list",
			tasks: []Task{},
			want:  true,
		},
		{
			name: "all done",
			tasks: []Task{
				{ID: 1, Status: TaskStatusDone},
				{ID: 2, Status: TaskStatusDone},
			},
			want: true,
		},
		{
			name: "all skipped",
			tasks: []Task{
				{ID: 1, Status: TaskStatusSkipped},
				{ID: 2, Status: TaskStatusSkipped},
			},
			want: true,
		},
		{
			name: "mixed done and skipped",
			tasks: []Task{
				{ID: 1, Status: TaskStatusDone},
				{ID: 2, Status: TaskStatusSkipped},
			},
			want: true,
		},
		{
			name: "with pending task",
			tasks: []Task{
				{ID: 1, Status: TaskStatusDone},
				{ID: 2, Status: TaskStatusPending},
			},
			want: false,
		},
		{
			name: "with in_progress task",
			tasks: []Task{
				{ID: 1, Status: TaskStatusInProgress},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			change := &Change{Tasks: tt.tasks}
			if got := AllTasksDone(change); got != tt.want {
				t.Errorf("AllTasksDone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAutoTransitionOnVerify(t *testing.T) {
	tests := []struct {
		name       string
		status     ChangeStatus
		passed     bool
		wantStatus ChangeStatus
		wantErr    bool
	}{
		{name: "passed in verifying", status: ChangeStatusVerifying, passed: true, wantStatus: ChangeStatusDone, wantErr: false},
		{name: "failed in verifying", status: ChangeStatusVerifying, passed: false, wantStatus: ChangeStatusImplementing, wantErr: false},
		{name: "auto-transition from draft", status: ChangeStatusDraft, passed: true, wantStatus: ChangeStatusDraft, wantErr: true},
		{name: "auto-transition from assigned", status: ChangeStatusAssigned, passed: true, wantStatus: ChangeStatusAssigned, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			change := &Change{Status: tt.status}
			err := AutoTransitionOnVerify(change, tt.passed)

			if (err != nil) != tt.wantErr {
				t.Errorf("AutoTransitionOnVerify() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && change.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", change.Status, tt.wantStatus)
			}
		})
	}
}
