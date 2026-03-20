package governance

import (
	"fmt"
	"time"
)

var validWorkflowPlanTransitions = map[string][]string{
	WorkflowPlanStatusProposed: {WorkflowPlanStatusApproved},
	WorkflowPlanStatusApproved: {WorkflowPlanStatusActive},
	WorkflowPlanStatusActive:   {WorkflowPlanStatusClosed},
	WorkflowPlanStatusClosed:   {},
}

func ValidateWorkflowPlanTransition(from, to string) error {
	allowed, ok := validWorkflowPlanTransitions[from]
	if !ok {
		return fmt.Errorf("unknown workflow plan status: %s", from)
	}
	for _, candidate := range allowed {
		if candidate == to {
			return nil
		}
	}
	return fmt.Errorf("invalid workflow plan transition: %s -> %s", from, to)
}

func NewWorkflowPlan(planID, taskID, owner string, inputRefs, reasons []string, now time.Time) *WorkflowPlan {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return &WorkflowPlan{
		ID:        planID,
		TaskID:    taskID,
		Owner:     owner,
		Status:    WorkflowPlanStatusProposed,
		InputRefs: append([]string(nil), inputRefs...),
		Reasons:   append([]string(nil), reasons...),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func ApproveWorkflowPlan(plan *WorkflowPlan, actor string, now time.Time) error {
	if plan == nil {
		return fmt.Errorf("workflow plan is nil")
	}
	if actor != plan.Owner {
		return fmt.Errorf("owner signoff required: actor %q is not owner %q", actor, plan.Owner)
	}
	if err := ValidateWorkflowPlanTransition(plan.Status, WorkflowPlanStatusApproved); err != nil {
		return err
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	plan.Status = WorkflowPlanStatusApproved
	plan.UpdatedAt = now
	return nil
}

func ActivateWorkflowPlan(plan *WorkflowPlan, now time.Time) error {
	if plan == nil {
		return fmt.Errorf("workflow plan is nil")
	}
	if err := ValidateWorkflowPlanTransition(plan.Status, WorkflowPlanStatusActive); err != nil {
		return err
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	plan.Status = WorkflowPlanStatusActive
	plan.UpdatedAt = now
	return nil
}

func CloseWorkflowPlan(plan *WorkflowPlan, now time.Time) error {
	if plan == nil {
		return fmt.Errorf("workflow plan is nil")
	}
	if err := ValidateWorkflowPlanTransition(plan.Status, WorkflowPlanStatusClosed); err != nil {
		return err
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	plan.Status = WorkflowPlanStatusClosed
	plan.UpdatedAt = now
	return nil
}
