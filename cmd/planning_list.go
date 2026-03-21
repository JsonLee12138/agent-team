package cmd

import (
	"fmt"
	"strings"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newPlanningListCmd() *cobra.Command {
	var kind string
	var lifecycle string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List planning artifacts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunPlanningList(kind, lifecycle)
		},
	}
	cmd.Flags().StringVar(&kind, "kind", "", "Filter by kind: roadmap, milestone, or phase")
	cmd.Flags().StringVar(&lifecycle, "lifecycle", "", "Filter by lifecycle: planning, archived, or deprecated")
	return cmd
}

func (a *App) RunPlanningList(kindRaw, lifecycleRaw string) error {
	var kind internal.PlanningKind
	var lifecycle internal.PlanningLifecycle
	var err error
	if strings.TrimSpace(kindRaw) != "" {
		kind, err = internal.ParsePlanningKind(kindRaw)
		if err != nil {
			return err
		}
	}
	if strings.TrimSpace(lifecycleRaw) != "" {
		lifecycle, err = internal.ParsePlanningLifecycle(lifecycleRaw)
		if err != nil {
			return err
		}
	}
	records, err := internal.ListPlanningRecords(a.Git.Root(), kind, lifecycle)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		fmt.Println("No planning artifacts found.")
		return nil
	}
	fmt.Printf("%-40s %-10s %-12s %s\n", "ID", "Kind", "Lifecycle", "Title")
	fmt.Printf("%-40s %-10s %-12s %s\n", "────────────────────────────────────────", "──────────", "────────────", "────────────────────────")
	for _, record := range records {
		fmt.Printf("%-40s %-10s %-12s %s\n", record.ID, record.Kind, record.Lifecycle, record.Title)
	}
	return nil
}
