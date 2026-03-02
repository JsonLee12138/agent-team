// cmd/task_show.go
package cmd

import (
	"fmt"
	"os"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newTaskShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <worker-id> <change-name>",
		Short: "Show details of a task change",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunTaskShow(args[0], args[1])
		},
	}

	return cmd
}

func (a *App) RunTaskShow(workerID, changeName string) error {
	root := a.Git.Root()
	wtPath := internal.WtPath(root, a.WtBase, workerID)

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("worker '%s' not found", workerID)
	}

	change, err := internal.LoadChange(wtPath, changeName)
	if err != nil {
		return fmt.Errorf("load change: %w", err)
	}

	// Display change details
	fmt.Printf("Change: %s\n", change.Name)
	fmt.Printf("Description: %s\n", change.Description)
	fmt.Printf("Status: %s\n", change.Status)
	fmt.Printf("Created: %s\n", change.CreatedAt)
	if change.AssignedTo != "" {
		fmt.Printf("Assigned To: %s\n", change.AssignedTo)
	}

	// Display verify config
	if change.Verify.Command != "" {
		fmt.Printf("Verify Command: %s\n", change.Verify.Command)
		fmt.Printf("Verify Timeout: %s\n", change.Verify.Timeout)
	}
	if change.Verify.Skip {
		fmt.Printf("Verify: skipped\n")
	}

	// Display tasks
	if len(change.Tasks) > 0 {
		fmt.Printf("\nTasks:\n")
		for _, task := range change.Tasks {
			fmt.Printf("  [%d] %s (%s)\n", task.ID, task.Title, task.Status)
		}
	}

	// Display file list
	fmt.Printf("\nFiles:\n")
	changeDir := internal.ChangeDirPath(wtPath, changeName)
	entries, err := os.ReadDir(changeDir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				fmt.Printf("  - %s\n", e.Name())
			}
		}
	}

	return nil
}
