// cmd/skill.go
package cmd

import "github.com/spf13/cobra"

func newSkillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage project skill cache",
	}
	cmd.AddCommand(newSkillCheckCmd())
	cmd.AddCommand(newSkillUpdateCmd())
	cmd.AddCommand(newSkillCleanCmd())
	return cmd
}
