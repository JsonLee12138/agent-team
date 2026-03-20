package orchestrator

import (
	"fmt"
	"time"

	"github.com/JsonLee12138/agent-team/internal/governance"
)

type GenerateWorkflowPlanInput struct {
	PlanID             string
	TaskID             string
	Owner              string
	ModuleID           string
	DeclaredReferences []string
	EvidenceRefs       []string
	Reasons            []string
	UsesArchivedInput  bool
	PublicRules        []governance.Rule
	ModuleRules        []governance.Rule
	TaskRules          []governance.Rule
	ArchivedTicket     *governance.ArchivedExceptionTicket
	Now                time.Time
}

func (u *Usecases) GenerateWorkflowPlan(input GenerateWorkflowPlanInput) (*governance.WorkflowPlan, error) {
	packet := governance.TaskPacket{
		TaskID:             input.TaskID,
		ModuleID:           input.ModuleID,
		Owner:              input.Owner,
		DeclaredReferences: append([]string(nil), input.DeclaredReferences...),
		UsesArchivedInput:  input.UsesArchivedInput,
	}

	gateResult, err := u.EvaluateGate(GateCheckInput{
		TaskPacket:     packet,
		PublicRules:    input.PublicRules,
		ModuleRules:    input.ModuleRules,
		TaskRules:      input.TaskRules,
		ArchivedTicket: input.ArchivedTicket,
	})
	if err != nil {
		return nil, err
	}
	if err := gateResultError(gateResult); err != nil {
		return nil, err
	}

	plan, err := governance.GenerateWorkflowPlan(governance.AdvisorInput{
		PlanID:       input.PlanID,
		TaskPacket:   packet,
		EvidenceRefs: append([]string(nil), input.EvidenceRefs...),
		Reasons:      append([]string(nil), input.Reasons...),
		Now:          input.Now,
	})
	if err != nil {
		return nil, fmt.Errorf("generate workflow plan: %w", err)
	}

	if err := u.Workflow.SaveWorkflowPlan(plan); err != nil {
		return nil, fmt.Errorf("save workflow plan: %w", err)
	}
	return plan, nil
}
