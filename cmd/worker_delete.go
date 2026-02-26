// cmd/worker_delete.go
package cmd

import (
	"fmt"
	"os"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newWorkerDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <worker-id>",
		Short: "Remove a worker and its worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunWorkerDelete(args[0])
		},
	}
}

func (a *App) RunWorkerDelete(workerID string) error {
	root := a.Git.Root()
	wtPath := internal.WtPath(root, a.WtBase, workerID)
	configPath := internal.WorkerConfigPath(root, workerID)
	workerDir := internal.WorkerDir(root, workerID)

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("worker '%s' not found", workerID)
	}

	fmt.Printf("Deleting worker '%s'...\n", workerID)

	// Kill running pane if any
	cfg, err := internal.LoadWorkerConfig(configPath)
	if err == nil && a.Session.PaneAlive(cfg.PaneID) {
		if killErr := a.Session.KillPane(cfg.PaneID); killErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close pane %s; continuing delete\n", cfg.PaneID)
		}
	}

	// Remove worktree
	if err := a.Git.WorktreeRemove(wtPath); err != nil {
		os.RemoveAll(wtPath)
	}

	// Delete branch
	a.Git.DeleteBranch("team/" + workerID)

	// Remove worker config directory
	os.RemoveAll(workerDir)

	fmt.Printf("✓ Deleted worker '%s'\n", workerID)
	return nil
}
