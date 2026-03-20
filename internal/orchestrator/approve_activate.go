package orchestrator

import (
	"fmt"
	"time"

	"github.com/JsonLee12138/agent-team/internal/governance"
)

type ApproveWorkflowPlanInput struct {
	PlanID string
	Actor  string
	Now    time.Time
}

func (u *Usecases) ApproveWorkflowPlan(input ApproveWorkflowPlanInput) (*governance.WorkflowPlan, error) {
	plan, err := u.Workflow.LoadWorkflowPlan(input.PlanID)
	if err != nil {
		return nil, fmt.Errorf("load workflow plan: %w", err)
	}

	gateResult, err := u.EvaluateGate(GateCheckInput{
		TaskPacket: governance.TaskPacket{
			TaskID:  plan.TaskID,
			ModuleID: "workflow",
			Owner:   plan.Owner,
		},
	})
	if err != nil {
		return nil, err
	}
	if err := gateResultError(gateResult); err != nil {
		return nil, err
	}

	if err := governance.ApproveWorkflowPlan(plan, input.Actor, input.Now); err != nil {
		return nil, err
	}
	if err := u.Workflow.SaveWorkflowPlan(plan); err != nil {
		return nil, fmt.Errorf("save workflow plan: %w", err)
	}
	return plan, nil
}

type ActivateWorkflowPlanInput struct {
	PlanID string
	Now    time.Time
}

func (u *Usecases) ActivateWorkflowPlan(input ActivateWorkflowPlanInput) (*governance.WorkflowPlan, error) {
	plan, err := u.Workflow.LoadWorkflowPlan(input.PlanID)
	if err != nil {
		return nil, fmt.Errorf("load workflow plan: %w", err)
	}

	gateResult, err := u.EvaluateGate(GateCheckInput{
		TaskPacket: governance.TaskPacket{
			TaskID:  plan.TaskID,
			ModuleID: "workflow",
			Owner:   plan.Owner,
		},
	})
	if err != nil {
		return nil, err
	}
	if err := gateResultError(gateResult); err != nil {
		return nil, err
	}

	if err := governance.ActivateWorkflowPlan(plan, input.Now); err != nil {
		return nil, err
	}
	if err := u.Workflow.SaveWorkflowPlan(plan); err != nil {
		return nil, fmt.Errorf("save workflow plan: %w", err)
	}
	return plan, nil
}

type CloseWorkflowPlanInput struct {
	PlanID string
	Now    time.Time
}

func (u *Usecases) CloseWorkflowPlan(input CloseWorkflowPlanInput) (*governance.WorkflowPlan, error) {
	plan, err := u.Workflow.LoadWorkflowPlan(input.PlanID)
	if err != nil {
		return nil, fmt.Errorf("load workflow plan: %w", err)
	}

	if err := governance.CloseWorkflowPlan(plan, input.Now); err != nil {
		return nil, err
	}
	if err := u.Workflow.SaveWorkflowPlan(plan); err != nil {
		return nil, fmt.Errorf("save workflow plan: %w", err)
	}
	return plan, nil
}
