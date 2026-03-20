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
	configDir := internal.WorkerConfigDir(root, workerID)
	configPath := internal.WorkerConfigPath(root, workerID)

	_, wtErr := os.Stat(wtPath)
	_, cfgErr := os.Stat(configPath)
	if os.IsNotExist(wtErr) && os.IsNotExist(cfgErr) {
		return fmt.Errorf("worker '%s' not found", workerID)
	}

	fmt.Printf("Deleting worker '%s'...\n", workerID)

	cfg, loadedConfigPath, err := internal.LoadWorkerConfigByID(root, a.WtBase, workerID)
	if err == nil && !os.IsNotExist(wtErr) {
		if err := closeWorkerSession(a.Session, cfg, loadedConfigPath); err != nil {
			return fmt.Errorf("failed to close session before delete: %w", err)
		}
	}

	if !os.IsNotExist(wtErr) {
		if err := a.Git.WorktreeRemove(wtPath); err != nil {
			os.RemoveAll(wtPath)
		}
	}

	a.Git.DeleteBranch("team/" + workerID)
	_ = os.Remove(configPath)
	_ = os.RemoveAll(configDir)

	fmt.Printf("✓ Deleted worker '%s'\n", workerID)
	return nil
}
