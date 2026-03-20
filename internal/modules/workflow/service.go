package workflow

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/JsonLee12138/agent-team/internal/governance"
	"gopkg.in/yaml.v3"
)

type Service struct {
	Root string
}

func NewService(root string) Service {
	return Service{Root: root}
}

func (s Service) NewTemplate(name, preset, ctoRole, devRole, qaRole, executionMode string) (*internal.WorkflowTemplate, error) {
	return internal.NewWorkflowTemplate(name, preset, ctoRole, devRole, qaRole, executionMode)
}

func (s Service) SaveTemplate(path string, wf *internal.WorkflowTemplate) error {
	return internal.SaveWorkflowTemplate(path, wf)
}

func (s Service) LoadTemplate(path string) (*internal.WorkflowTemplate, error) {
	return internal.LoadWorkflowTemplate(path)
}

func (s Service) ValidateTemplate(wf *internal.WorkflowTemplate) []string {
	return internal.ValidateWorkflowTemplate(wf)
}

func (s Service) NewRunState(workflowFile string, wf *internal.WorkflowTemplate, runID string) *internal.WorkflowRunState {
	return internal.NewWorkflowRunState(workflowFile, wf, runID)
}

func (s Service) LoadRunState(path string) (*internal.WorkflowRunState, error) {
	return internal.LoadWorkflowRunState(path)
}

func (s Service) WorkflowRunPath(workflowName, runID string) string {
	return internal.WorkflowRunPath(s.Root, workflowName, runID)
}

func (s Service) WorkflowTemplatePath(name string) string {
	return internal.WorkflowTemplatePath(s.Root, name)
}

func (s Service) SaveWorkflowPlan(plan *governance.WorkflowPlan) error {
	if plan == nil {
		return fmt.Errorf("workflow plan is nil")
	}
	path := s.workflowPlanPath(plan.ID)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create workflow plan directory: %w", err)
	}
	data, err := yaml.Marshal(plan)
	if err != nil {
		return fmt.Errorf("marshal workflow plan: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write workflow plan: %w", err)
	}
	return nil
}

func (s Service) LoadWorkflowPlan(planID string) (*governance.WorkflowPlan, error) {
	path := s.workflowPlanPath(planID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow plan: %w", err)
	}
	var plan governance.WorkflowPlan
	if err := yaml.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("parse workflow plan: %w", err)
	}
	return &plan, nil
}

func (s Service) workflowPlanPath(planID string) string {
	return filepath.Join(internal.WorkflowDir(s.Root), "plans", fmt.Sprintf("%s.yaml", planID))
}
