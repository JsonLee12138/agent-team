// cmd/worker_close.go
package cmd

import (
	"fmt"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newWorkerCloseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "close <worker-id>",
		Short: "Close a worker session without deleting the worker",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunWorkerClose(args[0])
		},
	}
}

func (a *App) RunWorkerClose(workerID string) error {
	root := a.Git.Root()
	wtPath := internal.WtPath(root, a.WtBase, workerID)
	configPath := internal.WorkerYAMLPath(wtPath)

	cfg, err := internal.LoadWorkerConfig(configPath)
	if err != nil {
		return fmt.Errorf("worker '%s' not found: %w", workerID, err)
	}

	if err := closeWorkerSession(a.Session, cfg, configPath); err != nil {
		return err
	}

	fmt.Printf("✓ Closed worker '%s'\n", workerID)
	return nil
}

// closeWorkerSession is a shared helper that shuts down the terminal pane
// for a worker and clears PaneID. It is idempotent: it succeeds if the pane
// is already gone or PaneID is already empty. It preserves ControllerPaneID,
// Provider, and DefaultModel.
func closeWorkerSession(session internal.SessionBackend, cfg *internal.WorkerConfig, configPath string) error {
	// Already closed — nothing to do.
	if cfg.PaneID == "" {
		return nil
	}

	// Pane is set but already offline — normalize state.
	if !session.PaneAlive(cfg.PaneID) {
		cfg.PaneID = ""
		return cfg.Save(configPath)
	}

	// Pane is alive — kill it.
	if err := session.KillPane(cfg.PaneID); err != nil {
		return fmt.Errorf("failed to close pane %s: %w", cfg.PaneID, err)
	}

	cfg.PaneID = ""
	return cfg.Save(configPath)
}
