// cmd/req.go
package cmd

import (
	"github.com/spf13/cobra"
)

func newReqCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "req",
		Short: "Manage requirements and sub-tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newReqCreateCmd())
	cmd.AddCommand(newReqSplitCmd())
	cmd.AddCommand(newReqAssignCmd())
	cmd.AddCommand(newReqStatusCmd())
	cmd.AddCommand(newReqDoneCmd())

	return cmd
}
