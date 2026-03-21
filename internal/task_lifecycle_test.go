package internal

import "testing"

func TestValidateTaskTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    TaskStatus
		to      TaskStatus
		wantErr bool
	}{
		{name: "draft to assigned", from: TaskStatusDraft, to: TaskStatusAssigned},
		{name: "assigned to done", from: TaskStatusAssigned, to: TaskStatusDone},
		{name: "done to assigned", from: TaskStatusDone, to: TaskStatusAssigned},
		{name: "done to archived", from: TaskStatusDone, to: TaskStatusArchived},
		{name: "draft to done invalid", from: TaskStatusDraft, to: TaskStatusDone, wantErr: true},
		{name: "archived terminal", from: TaskStatusArchived, to: TaskStatusDone, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTaskTransition(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateTaskTransition(%s,%s) error=%v wantErr=%v", tt.from, tt.to, err, tt.wantErr)
			}
		})
	}
}
