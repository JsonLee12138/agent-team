// internal/requirement_test.go
package internal

import (
	"testing"
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
