package cmd

import "github.com/spf13/cobra"

func newPlanningCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "planning",
		Short: "Manage roadmap, milestone, and phase planning artifacts",
	}
	cmd.AddCommand(newPlanningCreateCmd())
	cmd.AddCommand(newPlanningListCmd())
	cmd.AddCommand(newPlanningShowCmd())
	cmd.AddCommand(newPlanningMoveCmd())
	return cmd
}
