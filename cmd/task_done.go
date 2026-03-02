// cmd/task_done.go
package cmd

import (
	"fmt"
	"os"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newTaskDoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done <worker-id> <change-name> <task-id>",
		Short: "Mark a task as done",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := 0
			_, err := fmt.Sscanf(args[2], "%d", &taskID)
			if err != nil {
				return fmt.Errorf("invalid task ID: %s", args[2])
			}
			return GetApp(cmd).RunTaskDone(args[0], args[1], taskID)
		},
	}

	return cmd
}

func (a *App) RunTaskDone(workerID, changeName string, taskID int) error {
	root := a.Git.Root()
	wtPath := internal.WtPath(root, a.WtBase, workerID)

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("worker '%s' not found", workerID)
	}

	change, err := internal.LoadChange(wtPath, changeName)
	if err != nil {
		return fmt.Errorf("load change: %w", err)
	}

	// Find the task
	var task *internal.Task
	for i := range change.Tasks {
		if change.Tasks[i].ID == taskID {
			task = &change.Tasks[i]
			break
		}
	}

	if task == nil {
		return fmt.Errorf("task %d not found in change '%s'", taskID, changeName)
	}

	// Mark as done
	if task.Status == internal.TaskStatusDone {
		fmt.Printf("✓ Task %d is already done\n", taskID)
		return nil
	}

	task.Status = internal.TaskStatusDone
	fmt.Printf("✓ Task %d marked as done: %s\n", taskID, task.Title)

	// Save change
	if err := internal.SaveChange(wtPath, change); err != nil {
		return fmt.Errorf("save change: %w", err)
	}

	// Check if all tasks are done
	if internal.AllTasksDone(change) {
		fmt.Printf("✓ All tasks done. Run: agent-team task verify %s %s\n", workerID, changeName)
	}

	return nil
}
