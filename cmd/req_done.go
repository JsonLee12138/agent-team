// cmd/req_done.go
package cmd

import (
	"fmt"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newReqDoneCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "done <name>",
		Short: "Mark a requirement as done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunReqDone(args[0], force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force completion even with pending sub-tasks")

	return cmd
}

func (a *App) RunReqDone(name string, force bool) error {
	root := a.Git.Root()

	req, err := internal.LoadRequirement(root, name)
	if err != nil {
		return fmt.Errorf("load requirement: %w", err)
	}

	if req.Status == internal.RequirementStatusDone {
		fmt.Printf("Requirement '%s' is already done.\n", name)
		return nil
	}

	// Check for pending sub-tasks
	var pending []internal.SubTask
	for _, st := range req.SubTasks {
		if st.Status != internal.SubTaskStatusDone && st.Status != internal.SubTaskStatusSkipped {
			pending = append(pending, st)
		}
	}

	if len(pending) > 0 && !force {
		fmt.Printf("Warning: %d sub-task(s) are not done:\n", len(pending))
		for _, st := range pending {
			fmt.Printf("  - #%d %s (%s)\n", st.ID, st.Title, st.Status)
		}
		return fmt.Errorf("use --force to mark as done anyway")
	}

	if len(pending) > 0 {
		fmt.Printf("Warning: forcing completion with %d pending sub-task(s)\n", len(pending))
	}

	req.Status = internal.RequirementStatusDone

	if err := internal.SaveRequirement(root, req); err != nil {
		return fmt.Errorf("save requirement: %w", err)
	}

	if err := internal.UpdateIndexEntry(root, req); err != nil {
		return fmt.Errorf("update index: %w", err)
	}

	fmt.Printf("Requirement '%s' marked as done.\n", name)
	return nil
}
