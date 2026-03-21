package cmd

import (
	"fmt"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newPlanningCreateCmd() *cobra.Command {
	var kind string
	cmd := &cobra.Command{
		Use:   `create --kind <roadmap|milestone|phase> "<title>"`,
		Short: "Create a planning artifact",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunPlanningCreate(kind, args[0])
		},
	}
	cmd.Flags().StringVar(&kind, "kind", "", "Planning kind: roadmap, milestone, or phase")
	_ = cmd.MarkFlagRequired("kind")
	return cmd
}

func (a *App) RunPlanningCreate(kindRaw, title string) error {
	kind, err := internal.ParsePlanningKind(kindRaw)
	if err != nil {
		return err
	}
	record, err := internal.CreatePlanningRecord(a.Git.Root(), kind, title, time.Now().UTC())
	if err != nil {
		return err
	}
	fmt.Printf("✓ Created %s '%s'\n", record.Kind, record.ID)
	fmt.Printf("  → Title: %s\n", record.Title)
	fmt.Printf("  → Path: %s\n", record.Path)
	fmt.Printf("  → Lifecycle: %s\n", record.Lifecycle)
	return nil
}
