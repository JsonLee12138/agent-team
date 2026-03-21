package cmd

import (
	"fmt"
	"strings"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newPlanningShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show a planning artifact",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunPlanningShow(args[0])
		},
	}
	return cmd
}

func (a *App) RunPlanningShow(id string) error {
	record, err := internal.LoadPlanningRecord(a.Git.Root(), id)
	if err != nil {
		return err
	}
	issues := internal.ValidatePlanningReferences(a.Git.Root(), record)

	fmt.Printf("ID: %s\n", record.ID)
	fmt.Printf("Kind: %s\n", record.Kind)
	fmt.Printf("Title: %s\n", record.Title)
	fmt.Printf("Status: %s\n", record.Status)
	fmt.Printf("Lifecycle: %s\n", record.Lifecycle)
	fmt.Printf("Path: %s\n", record.Path)
	fmt.Printf("Created At: %s\n", record.CreatedAt)
	fmt.Printf("Updated At: %s\n", record.UpdatedAt)
	if record.Goal != "" {
		fmt.Printf("Goal: %s\n", record.Goal)
	}
	if record.ArchivedAt != "" {
		fmt.Printf("Archived At: %s\n", record.ArchivedAt)
	}
	if record.DeprecatedAt != "" {
		fmt.Printf("Deprecated At: %s\n", record.DeprecatedAt)
	}
	if record.DeprecatedReason != "" {
		fmt.Printf("Deprecated Reason: %s\n", record.DeprecatedReason)
	}
	fmt.Printf("Roadmap IDs: %s\n", joinOrDash(record.RoadmapIDs))
	fmt.Printf("Milestone IDs: %s\n", joinOrDash(record.MilestoneIDs))
	fmt.Printf("Phase IDs: %s\n", joinOrDash(record.PhaseIDs))
	fmt.Printf("Task IDs: %s\n", joinOrDash(record.TaskIDs))
	if len(issues) == 0 {
		fmt.Println("Reference Checks: ok")
	} else {
		fmt.Println("Reference Checks:")
		for _, issue := range issues {
			fmt.Printf("- %s\n", issue)
		}
	}
	return nil
}

func joinOrDash(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}
