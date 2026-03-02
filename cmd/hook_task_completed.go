// cmd/hook_task_completed.go
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newHookTaskCompletedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "task-completed",
		Short: "Handle task completed event (archive change + notify main)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHookTaskCompleted(cmd)
		},
	}
}

func runHookTaskCompleted(cmd *cobra.Command) error {
	input, err := internal.ParseHookInput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] task-completed: parse input: %v\n", err)
		return nil
	}

	wt, err := internal.ResolveWorktree(input.CWD)
	if err != nil || wt == nil {
		return nil
	}

	// Find first active change
	active, err := internal.ListActiveChanges(wt.WtPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] task-completed: list changes: %v\n", err)
		return nil
	}

	if len(active) == 0 {
		fmt.Fprintf(os.Stderr, "[agent-team] task-completed: no active changes to archive\n")
		return nil
	}

	change := active[0]

	// Transition to archived using internal functions directly (avoids CLI arg mismatch)
	if err := internal.ApplyChangeTransition(change, internal.ChangeStatusArchived); err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] task-completed: transition failed for '%s': %v\n", change.Name, err)
		return nil
	}

	if err := internal.SaveChange(wt.WtPath, change); err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] task-completed: save change '%s': %v\n", change.Name, err)
		return nil
	}

	fmt.Fprintf(os.Stderr, "[agent-team] task-completed: archived change '%s'\n", change.Name)

	// Notify main controller
	msg := fmt.Sprintf("Task completed: change '%s' archived by worker '%s'", change.Name, wt.WorkerID)
	notifyCmd := exec.Command("agent-team", "reply-main", msg)
	notifyCmd.Dir = wt.WtPath
	notifyCmd.Stderr = os.Stderr
	if err := notifyCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] task-completed: notify main failed: %v\n", err)
	}

	return nil
}
