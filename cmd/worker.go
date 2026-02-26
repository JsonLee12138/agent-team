// cmd/worker.go
package cmd

import "github.com/spf13/cobra"

func newWorkerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worker",
		Short: "Manage workers (role instances in isolated worktrees)",
	}
	cmd.AddCommand(newWorkerCreateCmd())
	cmd.AddCommand(newWorkerOpenCmd())
	cmd.AddCommand(newWorkerAssignCmd())
	cmd.AddCommand(newWorkerStatusCmd())
	cmd.AddCommand(newWorkerMergeCmd())
	cmd.AddCommand(newWorkerDeleteCmd())
	return cmd
}
