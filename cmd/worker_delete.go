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
	configPath := internal.WorkerYAMLPath(wtPath)

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("worker '%s' not found", workerID)
	}

	fmt.Printf("Deleting worker '%s'...\n", workerID)

	// Close running session via shared helper
	cfg, err := internal.LoadWorkerConfig(configPath)
	if err == nil {
		if err := closeWorkerSession(a.Session, cfg, configPath); err != nil {
			return fmt.Errorf("failed to close session before delete: %w", err)
		}
	}

	// Remove worktree
	if err := a.Git.WorktreeRemove(wtPath); err != nil {
		os.RemoveAll(wtPath)
	}

	// Delete branch
	a.Git.DeleteBranch("team/" + workerID)

	fmt.Printf("✓ Deleted worker '%s'\n", workerID)
	return nil
}
