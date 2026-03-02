// cmd/task.go
package cmd

import (
	"github.com/spf13/cobra"
)

func newTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage task changes and verification",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newTaskCreateCmd())
	cmd.AddCommand(newTaskListCmd())
	cmd.AddCommand(newTaskShowCmd())
	cmd.AddCommand(newTaskVerifyCmd())
	cmd.AddCommand(newTaskArchiveCmd())
	cmd.AddCommand(newTaskDoneCmd())

	return cmd
}
