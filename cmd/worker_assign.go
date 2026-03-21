// cmd/worker_assign.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newWorkerAssignCmd() *cobra.Command {
	var model string
	var newWindow bool
	cmd := &cobra.Command{
		Use:   `assign <worker-id> <task-id> [provider]`,
		Short: "Compatibility wrapper for task assign",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := ""
			if len(args) > 2 {
				provider = args[2]
			}
			return GetApp(cmd).RunWorkerAssign(args[0], args[1], provider, model, newWindow)
		},
	}
	cmd.Flags().StringVarP(&model, "model", "m", "", "AI model identifier")
	cmd.Flags().BoolVarP(&newWindow, "new-window", "w", false, "Open in a new window instead of a tab")
	return cmd
}

func (a *App) RunWorkerAssign(workerID, taskID, provider, model string, newWindow bool) error {
	fmt.Println("worker assign is deprecated; delegating to 'agent-team task assign'.")
	return a.RunTaskAssign(taskID, workerID, provider, model, newWindow)
}
