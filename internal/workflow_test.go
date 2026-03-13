package internal

import "testing"

func TestNewWorkflowTemplateBranching(t *testing.T) {
	t.Parallel()

	wf, err := NewWorkflowTemplate("feature-delivery", "branching", "cto", "vite-react-dev", "qa", "semi_auto")
	if err != nil {
		t.Fatalf("NewWorkflowTemplate() error = %v", err)
	}
	if wf.Entry != "cto_breakdown" {
		t.Fatalf("expected entry cto_breakdown, got %s", wf.Entry)
	}
	if len(ValidateWorkflowTemplate(wf)) != 0 {
		t.Fatalf("expected valid workflow template")
	}
}

func TestWorkflowRunStateTransitions(t *testing.T) {
	t.Parallel()

	wf, err := NewWorkflowTemplate("feature-delivery", "branching", "cto", "vite-react-dev", "qa", "semi_auto")
	if err != nil {
		t.Fatalf("NewWorkflowTemplate() error = %v", err)
	}
	state := NewWorkflowRunState(".agents/workflow/feature-delivery.yaml", wf, "run-1")

	if err := state.Start(wf, "cto_breakdown", false); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if state.Status != WorkflowStatusRunning {
		t.Fatalf("expected running status, got %s", state.Status)
	}

	nextNode, err := state.Complete(wf, "cto_breakdown", "", "", "cto-001", "done", "complete")
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if nextNode != "controller_dispatch" {
		t.Fatalf("expected controller_dispatch, got %s", nextNode)
	}
	if state.RoleWorkerMap["cto"] != "cto-001" {
		t.Fatalf("expected role worker map to be updated")
	}

	nextNode, err = state.Confirm(wf, "controller_dispatch", "dev_first", "", "", "branch chosen")
	if err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}
	if nextNode != "dev_implement" {
		t.Fatalf("expected dev_implement, got %s", nextNode)
	}
	if state.CurrentNode != "dev_implement" {
		t.Fatalf("expected current node dev_implement, got %s", state.CurrentNode)
	}
}
