package cmd

import (
	"fmt"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newPlanningMoveCmd() *cobra.Command {
	var to string
	var reason string
	cmd := &cobra.Command{
		Use:   "move <id> --to <planning|archived|deprecated>",
		Short: "Move a planning artifact between planning, archived, and deprecated",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunPlanningMove(args[0], to, reason)
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "Target lifecycle: planning, archived, or deprecated")
	cmd.Flags().StringVar(&reason, "reason", "", "Deprecated reason when moving to deprecated")
	_ = cmd.MarkFlagRequired("to")
	return cmd
}

func (a *App) RunPlanningMove(id, toRaw, reason string) error {
	to, err := internal.ParsePlanningLifecycle(toRaw)
	if err != nil {
		return err
	}
	record, err := internal.MovePlanningRecord(a.Git.Root(), id, to, reason, time.Now().UTC())
	if err != nil {
		return err
	}
	fmt.Printf("✓ Moved %s '%s' to %s\n", record.Kind, record.ID, record.Lifecycle)
	fmt.Printf("  → Path: %s\n", record.Path)
	return nil
}
