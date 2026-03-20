package governance

import (
	"testing"
	"time"
)

func TestGenerateWorkflowPlan_FromTextArtifactsOnly(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	plan, err := GenerateWorkflowPlan(AdvisorInput{
		PlanID: "plan-1",
		TaskPacket: TaskPacket{
			TaskID:             "task-1",
			Owner:              "owner-1",
			DeclaredReferences: []string{"req-1"},
		},
		EvidenceRefs: []string{"evidence-1"},
		Reasons:      []string{"text-based governance input"},
		Now:          now,
	})
	if err != nil {
		t.Fatalf("GenerateWorkflowPlan failed: %v", err)
	}

	if plan.Status != WorkflowPlanStatusProposed {
		t.Fatalf("expected proposed status")
	}
	if len(plan.InputRefs) != 3 {
		t.Fatalf("expected 3 input refs, got %d", len(plan.InputRefs))
	}
	if len(plan.Reasons) == 0 {
		t.Fatalf("expected reasons")
	}
}
