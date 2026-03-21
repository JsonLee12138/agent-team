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
		{name: "assigned to verifying", from: TaskStatusAssigned, to: TaskStatusVerifying},
		{name: "verifying to assigned", from: TaskStatusVerifying, to: TaskStatusAssigned},
		{name: "verifying to archived", from: TaskStatusVerifying, to: TaskStatusArchived},
		{name: "draft to deprecated", from: TaskStatusDraft, to: TaskStatusDeprecated},
		{name: "draft to verifying invalid", from: TaskStatusDraft, to: TaskStatusVerifying, wantErr: true},
		{name: "archived terminal", from: TaskStatusArchived, to: TaskStatusVerifying, wantErr: true},
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
