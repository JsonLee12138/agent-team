package cmd

import "github.com/spf13/cobra"

func newTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage task packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newTaskCreateCmd())
	cmd.AddCommand(newTaskListCmd())
	cmd.AddCommand(newTaskShowCmd())
	cmd.AddCommand(newTaskAssignCmd())
	cmd.AddCommand(newTaskDoneCmd())
	cmd.AddCommand(newTaskArchiveCmd())
	cmd.AddCommand(newTaskDeprecatedCmd())
	return cmd
}
