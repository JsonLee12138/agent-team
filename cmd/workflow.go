package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/JsonLee12138/agent-team/internal/orchestrator"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newWorkflowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Create and validate controller workflow files and run-state",
	}
	cmd.AddCommand(newWorkflowCreateCmd())
	cmd.AddCommand(newWorkflowValidateCmd())
	cmd.AddCommand(newWorkflowPlanCmd())
	cmd.AddCommand(newWorkflowStateCmd())
	return cmd
}

func newWorkflowCreateCmd() *cobra.Command {
	var preset string
	var output string
	var ctoRole string
	var devRole string
	var qaRole string
	var executionMode string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a workflow template from a preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := GetApp(cmd)
			uc := orchestrator.NewUsecases(app.Git.Root())
			workflow, err := uc.Workflow.NewTemplate(args[0], preset, ctoRole, devRole, qaRole, executionMode)
			if err != nil {
				return err
			}
			target := output
			if target == "" {
				target = uc.Workflow.WorkflowTemplatePath(workflow.Name)
			}
			if err := uc.Workflow.SaveTemplate(target, workflow); err != nil {
				return err
			}
			fmt.Println(target)
			return nil
		},
	}

	cmd.Flags().StringVar(&preset, "preset", "branching", "Starter flow preset (branching|dev-first|test-first)")
	cmd.Flags().StringVar(&output, "output", "", "Output path. Defaults to .agents/workflow/<name>.yaml")
	cmd.Flags().StringVar(&ctoRole, "cto-role", "cto", "Role name for cto alias")
	cmd.Flags().StringVar(&devRole, "dev-role", "vite-react-dev", "Role name for dev alias")
	cmd.Flags().StringVar(&qaRole, "qa-role", "qa", "Role name for qa alias")
	cmd.Flags().StringVar(&executionMode, "execution-mode", "semi_auto", "Execution mode (semi_auto|manual)")
	return cmd
}

func newWorkflowValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <workflow-file>",
		Short: "Validate a workflow template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := GetApp(cmd)
			uc := orchestrator.NewUsecases(app.Git.Root())
			wf, err := uc.Workflow.LoadTemplate(args[0])
			if err != nil {
				return err
			}
			errs := uc.Workflow.ValidateTemplate(wf)
			if len(errs) > 0 {
				return fmt.Errorf("workflow validation failed:\n- %s", joinErrors(errs))
			}
			fmt.Printf("OK: workflow is valid: %s\n", args[0])
			return nil
		},
	}
}

func newWorkflowPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Manage governance workflow plans",
	}
	cmd.AddCommand(newWorkflowPlanGenerateCmd())
	cmd.AddCommand(newWorkflowPlanApproveCmd())
	cmd.AddCommand(newWorkflowPlanActivateCmd())
	return cmd
}

func newWorkflowPlanGenerateCmd() *cobra.Command {
	var planID string
	var taskID string
	var owner string
	var moduleID string
	var ref []string
	var evidence []string
	var reason []string
	var archived bool

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a governance workflow plan (proposed)",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := GetApp(cmd)
			uc := orchestrator.NewUsecases(app.Git.Root())
			plan, err := uc.GenerateWorkflowPlan(orchestrator.GenerateWorkflowPlanInput{
				PlanID:             planID,
				TaskID:             taskID,
				Owner:              owner,
				ModuleID:           moduleID,
				DeclaredReferences: ref,
				EvidenceRefs:       evidence,
				Reasons:            reason,
				UsesArchivedInput:  archived,
				Now:                time.Now().UTC(),
			})
			if err != nil {
				return err
			}
			fmt.Printf("plan_id=%s\n", plan.ID)
			fmt.Printf("status=%s\n", plan.Status)
			return nil
		},
	}

	cmd.Flags().StringVar(&planID, "plan-id", "", "Workflow plan id")
	cmd.Flags().StringVar(&taskID, "task-id", "", "Task id")
	cmd.Flags().StringVar(&owner, "owner", "", "Owner id")
	cmd.Flags().StringVar(&moduleID, "module", "workflow", "Module id")
	cmd.Flags().StringArrayVar(&ref, "ref", nil, "Declared reference id (repeatable)")
	cmd.Flags().StringArrayVar(&evidence, "evidence", nil, "Evidence reference (repeatable)")
	cmd.Flags().StringArrayVar(&reason, "reason", nil, "Reason (repeatable)")
	cmd.Flags().BoolVar(&archived, "use-archived", false, "Uses archived input")
	_ = cmd.MarkFlagRequired("plan-id")
	_ = cmd.MarkFlagRequired("task-id")
	_ = cmd.MarkFlagRequired("owner")
	return cmd
}

func newWorkflowPlanApproveCmd() *cobra.Command {
	var planID string
	var actor string

	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve a workflow plan (owner only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := GetApp(cmd)
			uc := orchestrator.NewUsecases(app.Git.Root())
			plan, err := uc.ApproveWorkflowPlan(orchestrator.ApproveWorkflowPlanInput{
				PlanID: planID,
				Actor:  actor,
				Now:    time.Now().UTC(),
			})
			if err != nil {
				return err
			}
			fmt.Printf("plan_id=%s\n", plan.ID)
			fmt.Printf("status=%s\n", plan.Status)
			return nil
		},
	}

	cmd.Flags().StringVar(&planID, "plan-id", "", "Workflow plan id")
	cmd.Flags().StringVar(&actor, "actor", "", "Approver actor id")
	_ = cmd.MarkFlagRequired("plan-id")
	_ = cmd.MarkFlagRequired("actor")
	return cmd
}

func newWorkflowPlanActivateCmd() *cobra.Command {
	var planID string

	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Activate an approved workflow plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := GetApp(cmd)
			uc := orchestrator.NewUsecases(app.Git.Root())
			plan, err := uc.ActivateWorkflowPlan(orchestrator.ActivateWorkflowPlanInput{
				PlanID: planID,
				Now:    time.Now().UTC(),
			})
			if err != nil {
				return err
			}
			fmt.Printf("plan_id=%s\n", plan.ID)
			fmt.Printf("status=%s\n", plan.Status)
			return nil
		},
	}

	cmd.Flags().StringVar(&planID, "plan-id", "", "Workflow plan id")
	_ = cmd.MarkFlagRequired("plan-id")
	return cmd
}

func newWorkflowStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Manage workflow run-state files",
	}
	cmd.AddCommand(newWorkflowStateInitCmd())
	cmd.AddCommand(newWorkflowStateShowCmd())
	cmd.AddCommand(newWorkflowStateStartCmd())
	cmd.AddCommand(newWorkflowStateWaitCmd())
	cmd.AddCommand(newWorkflowStateBlockCmd())
	cmd.AddCommand(newWorkflowStateCompleteCmd())
	cmd.AddCommand(newWorkflowStateConfirmCmd())
	return cmd
}

func newWorkflowStateInitCmd() *cobra.Command {
	var stateFile string
	var runID string

	cmd := &cobra.Command{
		Use:   "init <workflow-file>",
		Short: "Initialize run-state from a workflow template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := GetApp(cmd)
			uc := orchestrator.NewUsecases(app.Git.Root())
			wf, err := uc.Workflow.LoadTemplate(args[0])
			if err != nil {
				return err
			}
			errs := uc.Workflow.ValidateTemplate(wf)
			if len(errs) > 0 {
				return fmt.Errorf("workflow validation failed:\n- %s", joinErrors(errs))
			}
			actualRunID := runID
			if actualRunID == "" {
				actualRunID = internal.WorkflowRunID(wf.Name, time.Now())
			}
			target := stateFile
			if target == "" {
				target = uc.Workflow.WorkflowRunPath(wf.Name, actualRunID)
			}
			workflowPath := args[0]
			if !filepath.IsAbs(workflowPath) {
				workflowPath = filepath.Clean(workflowPath)
			}
			state := uc.Workflow.NewRunState(workflowPath, wf, actualRunID)
			if err := state.Save(target); err != nil {
				return err
			}
			fmt.Println(target)
			fmt.Printf("current_node=%s\n", state.CurrentNode)
			return nil
		},
	}

	cmd.Flags().StringVar(&stateFile, "state-file", "", "Output path for run-state YAML file")
	cmd.Flags().StringVar(&runID, "run-id", "", "Explicit run id")
	return cmd
}

func newWorkflowStateShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <state-file>",
		Short: "Print a run-state file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := internal.LoadWorkflowRunState(args[0])
			if err != nil {
				return err
			}
			data, err := yaml.Marshal(state)
			if err != nil {
				return fmt.Errorf("marshal workflow run state: %w", err)
			}
			fmt.Print(string(data))
			return nil
		},
	}
}

func newWorkflowStateStartCmd() *cobra.Command {
	var nodeID string
	var force bool
	cmd := &cobra.Command{
		Use:   "start <state-file>",
		Short: "Mark a workflow node as running",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			state, wf, err := loadWorkflowStatePair(args[0])
			if err != nil {
				return err
			}
			if err := state.Start(wf, nodeID, force); err != nil {
				return err
			}
			if err := state.Save(args[0]); err != nil {
				return err
			}
			fmt.Printf("started %s\n", nodeID)
			return nil
		},
	}
	cmd.Flags().StringVar(&nodeID, "node", "", "Node id to start")
	cmd.Flags().BoolVar(&force, "force", false, "Allow restarting a completed node")
	_ = cmd.MarkFlagRequired("node")
	return cmd
}

func newWorkflowStateWaitCmd() *cobra.Command {
	var nodeID string
	var reason string
	cmd := &cobra.Command{
		Use:   "wait <state-file>",
		Short: "Mark a workflow node as waiting",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			state, wf, err := loadWorkflowStatePair(args[0])
			if err != nil {
				return err
			}
			if err := state.Wait(wf, nodeID, reason); err != nil {
				return err
			}
			if err := state.Save(args[0]); err != nil {
				return err
			}
			fmt.Printf("waiting on %s\n", nodeID)
			return nil
		},
	}
	cmd.Flags().StringVar(&nodeID, "node", "", "Node id to mark waiting")
	cmd.Flags().StringVar(&reason, "reason", "", "Reason for waiting")
	_ = cmd.MarkFlagRequired("node")
	_ = cmd.MarkFlagRequired("reason")
	return cmd
}

func newWorkflowStateBlockCmd() *cobra.Command {
	var nodeID string
	var reason string
	cmd := &cobra.Command{
		Use:   "block <state-file>",
		Short: "Block a workflow run on a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			state, wf, err := loadWorkflowStatePair(args[0])
			if err != nil {
				return err
			}
			if err := state.Block(wf, nodeID, reason); err != nil {
				return err
			}
			if err := state.Save(args[0]); err != nil {
				return err
			}
			fmt.Printf("blocked on %s\n", nodeID)
			return nil
		},
	}
	cmd.Flags().StringVar(&nodeID, "node", "", "Node id to block")
	cmd.Flags().StringVar(&reason, "reason", "", "Blocking reason")
	_ = cmd.MarkFlagRequired("node")
	_ = cmd.MarkFlagRequired("reason")
	return cmd
}

func newWorkflowStateCompleteCmd() *cobra.Command {
	return newWorkflowStateAdvanceCmd("complete")
}

func newWorkflowStateConfirmCmd() *cobra.Command {
	return newWorkflowStateAdvanceCmd("confirm")
}

func newWorkflowStateAdvanceCmd(action string) *cobra.Command {
	var nodeID string
	var outcome string
	var nextNode string
	var workerID string
	var summary string

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s <state-file>", action),
		Short: fmt.Sprintf("%s a workflow node and advance", action),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			state, wf, err := loadWorkflowStatePair(args[0])
			if err != nil {
				return err
			}
			var resolvedNext string
			if action == "confirm" {
				resolvedNext, err = state.Confirm(wf, nodeID, outcome, nextNode, workerID, summary)
			} else {
				resolvedNext, err = state.Complete(wf, nodeID, outcome, nextNode, workerID, summary, action)
			}
			if err != nil {
				return err
			}
			if err := state.Save(args[0]); err != nil {
				return err
			}
			fmt.Printf("completed %s\n", nodeID)
			if resolvedNext == "" {
				fmt.Println("next_node=<end>")
			} else {
				fmt.Printf("next_node=%s\n", resolvedNext)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&nodeID, "node", "", "Node id to advance")
	cmd.Flags().StringVar(&outcome, "outcome", "", "Outcome label used for branch selection")
	cmd.Flags().StringVar(&nextNode, "next-node", "", "Explicit next node override")
	cmd.Flags().StringVar(&workerID, "worker", "", "Worker id associated with this node result")
	cmd.Flags().StringVar(&summary, "summary", "", "Optional summary for the node result")
	_ = cmd.MarkFlagRequired("node")
	return cmd
}

func loadWorkflowStatePair(stateFile string) (*internal.WorkflowRunState, *internal.WorkflowTemplate, error) {
	state, err := internal.LoadWorkflowRunState(stateFile)
	if err != nil {
		return nil, nil, err
	}
	wf, err := internal.LoadWorkflowTemplate(state.WorkflowFile)
	if err != nil {
		return nil, nil, err
	}
	errs := internal.ValidateWorkflowTemplate(wf)
	if len(errs) > 0 {
		return nil, nil, fmt.Errorf("workflow validation failed:\n- %s", joinErrors(errs))
	}
	return state, wf, nil
}

func joinErrors(errs []string) string {
	return strings.Join(errs, "\n- ")
}
