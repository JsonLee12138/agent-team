package orchestrator

import (
	"fmt"

	"github.com/JsonLee12138/agent-team/internal/governance"
	requirementmodule "github.com/JsonLee12138/agent-team/internal/modules/requirement"
	taskmodule "github.com/JsonLee12138/agent-team/internal/modules/task"
	workflowmodule "github.com/JsonLee12138/agent-team/internal/modules/workflow"
)

type Usecases struct {
	Root              string
	Task              taskmodule.Service
	Requirement       requirementmodule.Service
	Workflow          workflowmodule.Service
	ExceptionRegistry governance.ArchivedExceptionRegistry
}

func NewUsecases(root string) *Usecases {
	return &Usecases{
		Root:              root,
		Task:              taskmodule.NewService(root),
		Requirement:       requirementmodule.NewService(root),
		Workflow:          workflowmodule.NewService(root),
		ExceptionRegistry: governance.NewInMemoryArchivedExceptionRegistry(),
	}
}

type GateCheckInput struct {
	TaskPacket     governance.TaskPacket
	PublicRules    []governance.Rule
	ModuleRules    []governance.Rule
	TaskRules      []governance.Rule
	ArchivedTicket *governance.ArchivedExceptionTicket
}

func (u *Usecases) EvaluateGate(input GateCheckInput) (governance.GateResult, error) {
	index, err := u.Requirement.LoadGovernanceIndex()
	if err != nil {
		return governance.GateResult{}, fmt.Errorf("load governance index: %w", err)
	}

	loadedRules := governance.LoadRules(input.PublicRules, input.ModuleRules, input.TaskRules)
	result := governance.EvaluateGate(governance.GateInput{
		TaskPacket:     input.TaskPacket,
		Index:          index,
		LoadedRules:    loadedRules,
		ArchivedTicket: input.ArchivedTicket,
	})

	return result, nil
}

func gateResultError(result governance.GateResult) error {
	if !result.IsBlocker() {
		return nil
	}
	return fmt.Errorf("gate blocked: code=%s message=%s", result.Code, result.Message)
}
