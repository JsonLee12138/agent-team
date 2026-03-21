package cmd

import (
	"fmt"
	"time"

	"github.com/JsonLee12138/agent-team/internal/orchestrator"
	"github.com/spf13/cobra"
)

func newWorkflowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Manage governance workflow plans",
	}
	cmd.AddCommand(newWorkflowPlanCmd())
	return cmd
}

func newWorkflowPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Manage governance workflow plans",
	}
	cmd.AddCommand(newWorkflowPlanGenerateCmd())
	cmd.AddCommand(newWorkflowPlanApproveCmd())
	cmd.AddCommand(newWorkflowPlanActivateCmd())
	cmd.AddCommand(newWorkflowPlanCloseCmd())
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

func newWorkflowPlanCloseCmd() *cobra.Command {
	var planID string

	cmd := &cobra.Command{
		Use:   "close",
		Short: "Close an active workflow plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := GetApp(cmd)
			uc := orchestrator.NewUsecases(app.Git.Root())
			plan, err := uc.CloseWorkflowPlan(orchestrator.CloseWorkflowPlanInput{
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
