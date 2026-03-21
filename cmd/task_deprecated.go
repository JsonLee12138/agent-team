package cmd

import (
	"fmt"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newTaskDeprecatedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deprecated <task-id>",
		Short: "Move a task to deprecated state",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunTaskDeprecated(args[0])
		},
	}
	return cmd
}

func (a *App) RunTaskDeprecated(taskID string) error {
	record, err := internal.DeprecateTask(a.Git.Root(), taskID, time.Now().UTC())
	if err != nil {
		return err
	}
	fmt.Printf("✓ Deprecated task '%s'\n", record.TaskID)
	return nil
}
