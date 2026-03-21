package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newTaskCreateCmd() *cobra.Command {
	var role string
	var design string
	cmd := &cobra.Command{
		Use:   `create --role <role> "<title>"`,
		Short: "Create a task package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunTaskCreate(args[0], role, design)
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "Role bound to this task")
	cmd.Flags().StringVar(&design, "design", "", "Path to design/brainstorming file")
	_ = cmd.MarkFlagRequired("role")
	return cmd
}

func (a *App) RunTaskCreate(title, role, designPath string) error {
	root := a.Git.Root()
	if _, err := internal.ResolveRole(root, role); err != nil {
		return err
	}

	design := ""
	if designPath != "" {
		data, err := os.ReadFile(designPath)
		if err != nil {
			return fmt.Errorf("read design: %w", err)
		}
		design = string(data)
	}

	record, err := internal.CreateTaskPackage(root, title, role, design, time.Now().UTC())
	if err != nil {
		return err
	}

	fmt.Printf("✓ Created task '%s'\n", record.TaskID)
	fmt.Printf("  → Role: %s\n", record.Role)
	fmt.Printf("  → Path: %s\n", record.TaskPath)
	fmt.Printf("  → Status: %s\n", record.Status)
	return nil
}
