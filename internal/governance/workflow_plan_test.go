package governance

import (
	"testing"
	"time"
)

func TestWorkflowPlanTransitionsAndOwnerSignoff(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	plan := NewWorkflowPlan("plan-1", "task-1", "owner-1", []string{"task-1"}, []string{"reason"}, now)

	if plan.Status != WorkflowPlanStatusProposed {
		t.Fatalf("expected proposed status")
	}

	if err := ApproveWorkflowPlan(plan, "other", now); err == nil {
		t.Fatalf("expected owner-only approve guard")
	}

	if err := ApproveWorkflowPlan(plan, "owner-1", now); err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	if plan.Status != WorkflowPlanStatusApproved {
		t.Fatalf("expected approved")
	}

	if err := ActivateWorkflowPlan(plan, now); err != nil {
		t.Fatalf("activate failed: %v", err)
	}
	if plan.Status != WorkflowPlanStatusActive {
		t.Fatalf("expected active")
	}

	if err := CloseWorkflowPlan(plan, now); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if plan.Status != WorkflowPlanStatusClosed {
		t.Fatalf("expected closed")
	}
}
