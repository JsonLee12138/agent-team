package cmd

import (
	"fmt"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newTaskDoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done <task-id>",
		Short: "Mark an assigned task as done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunTaskDone(args[0])
		},
	}
	return cmd
}

func (a *App) RunTaskDone(taskID string) error {
	record, err := internal.MarkTaskDone(a.Git.Root(), taskID, time.Now().UTC())
	if err != nil {
		return err
	}
	fmt.Printf("✓ Task '%s' marked as %s\n", record.TaskID, record.Status)
	return nil
}
