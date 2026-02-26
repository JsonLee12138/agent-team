// cmd/worker_status.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newWorkerStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show all workers, their roles, running state, and active changes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunWorkerStatus()
		},
	}
}

func (a *App) RunWorkerStatus() error {
	root := a.Git.Root()
	workers := internal.ListWorkers(root)
	if len(workers) == 0 {
		fmt.Println("No workers found. Create one with: agent-team worker create <role-name>")
		return nil
	}

	fmt.Printf("%-24s %-16s %-24s %s\n", "Worker", "Role", "Status", "Changes")
	fmt.Printf("%-24s %-16s %-24s %s\n", "────────────────────────", "────────────────", "────────────────────────", "──────────────────────────")

	for _, w := range workers {
		status := "✗ offline"
		if w.Config != nil && a.Session.PaneAlive(w.Config.PaneID) {
			status = fmt.Sprintf("✓ running [p:%s]", w.Config.PaneID)
		}

		// Count changes
		wtPath := internal.WtPath(root, a.WtBase, w.WorkerID)
		changesDir := filepath.Join(wtPath, "openspec", "changes")
		changesSummary := "0"
		if entries, err := os.ReadDir(changesDir); err == nil {
			active := 0
			for _, e := range entries {
				if e.IsDir() && e.Name() != "archive" {
					active++
				}
			}
			if active > 0 {
				changesSummary = fmt.Sprintf("%d active", active)
			}
		}

		fmt.Printf("%-24s %-16s %-24s %s\n", w.WorkerID, w.Role, status, changesSummary)
	}
	return nil
}
