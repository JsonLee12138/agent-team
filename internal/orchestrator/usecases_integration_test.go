package orchestrator

import (
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/JsonLee12138/agent-team/internal/governance"
)

func TestUsecasesGenerateApproveActivate_MainPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := internal.SaveRequirementIndex(root, &internal.RequirementIndex{
		Requirements: []internal.RequirementIndexEntry{{
			Name:         "task-1",
			Status:       internal.RequirementStatusOpen,
			SubTaskCount: 0,
			DoneCount:    0,
		}},
	}); err != nil {
		t.Fatalf("SaveRequirementIndex failed: %v", err)
	}

	uc := NewUsecases(root)
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)

	plan, err := uc.GenerateWorkflowPlan(GenerateWorkflowPlanInput{
		PlanID:            "plan-1",
		TaskID:            "task-1",
		Owner:             "owner-1",
		ModuleID:          "workflow",
		EvidenceRefs:      []string{"evidence-1"},
		Reasons:           []string{"integration test"},
		PublicRules:       []governance.Rule{{ID: "p1", Key: "k1", Value: "v1"}},
		UsesArchivedInput: false,
		Now:               now,
	})
	if err != nil {
		t.Fatalf("GenerateWorkflowPlan failed: %v", err)
	}
	if plan.Status != governance.WorkflowPlanStatusProposed {
		t.Fatalf("expected proposed, got %s", plan.Status)
	}

	plan, err = uc.ApproveWorkflowPlan(ApproveWorkflowPlanInput{
		PlanID: "plan-1",
		Actor:  "owner-1",
		Now:    now,
	})
	if err != nil {
		t.Fatalf("ApproveWorkflowPlan failed: %v", err)
	}
	if plan.Status != governance.WorkflowPlanStatusApproved {
		t.Fatalf("expected approved, got %s", plan.Status)
	}

	plan, err = uc.ActivateWorkflowPlan(ActivateWorkflowPlanInput{
		PlanID: "plan-1",
		Now:    now,
	})
	if err != nil {
		t.Fatalf("ActivateWorkflowPlan failed: %v", err)
	}
	if plan.Status != governance.WorkflowPlanStatusActive {
		t.Fatalf("expected active, got %s", plan.Status)
	}
}

func TestUsecasesGenerateWorkflowPlan_BlockedWhenTaskNotIndexed(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := internal.SaveRequirementIndex(root, &internal.RequirementIndex{}); err != nil {
		t.Fatalf("SaveRequirementIndex failed: %v", err)
	}

	uc := NewUsecases(root)
	_, err := uc.GenerateWorkflowPlan(GenerateWorkflowPlanInput{
		PlanID:   "plan-1",
		TaskID:   "task-1",
		Owner:    "owner-1",
		ModuleID: "workflow",
	})
	if err == nil {
		t.Fatalf("expected gate blocker when task not indexed")
	}
}

func TestUsecasesApproveWorkflowPlan_RequireOwnerSignoff(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := internal.SaveRequirementIndex(root, &internal.RequirementIndex{
		Requirements: []internal.RequirementIndexEntry{{Name: "task-1", Status: internal.RequirementStatusOpen}},
	}); err != nil {
		t.Fatalf("SaveRequirementIndex failed: %v", err)
	}

	uc := NewUsecases(root)
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	if _, err := uc.GenerateWorkflowPlan(GenerateWorkflowPlanInput{
		PlanID: "plan-1",
		TaskID: "task-1",
		Owner:  "owner-1",
		Now:    now,
	}); err != nil {
		t.Fatalf("GenerateWorkflowPlan failed: %v", err)
	}

	if _, err := uc.ApproveWorkflowPlan(ApproveWorkflowPlanInput{
		PlanID: "plan-1",
		Actor:  "not-owner",
		Now:    now,
	}); err == nil {
		t.Fatalf("expected owner signoff error")
	}
}
