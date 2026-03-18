// cmd/req_create.go
package cmd

import (
	"fmt"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newReqCreateCmd() *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new requirement",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunReqCreate(args[0], description)
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Requirement description")

	return cmd
}

func (a *App) RunReqCreate(name, description string) error {
	root := a.Git.Root()

	req, err := internal.CreateRequirement(root, name, description, nil)
	if err != nil {
		return fmt.Errorf("create requirement: %w", err)
	}

	if err := internal.UpdateIndexEntry(root, req); err != nil {
		return fmt.Errorf("update index: %w", err)
	}

	fmt.Printf("Created requirement: %s\n", req.Name)
	fmt.Printf("  Status: %s\n", req.Status)
	return nil
}
