// cmd/reply.go
package cmd

import (
	"fmt"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newReplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   `reply <worker-id> "<answer>"`,
		Short: "Send a reply to a worker's running session",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunReply(args[0], args[1])
		},
	}
}

func (a *App) RunReply(workerID, answer string) error {
	root := a.Git.Root()

	// Try v2 worker config first
	configPath := internal.WorkerConfigPath(root, workerID)
	cfg, err := internal.LoadWorkerConfig(configPath)
	if err == nil {
		if !a.Session.PaneAlive(cfg.PaneID) {
			return fmt.Errorf("worker '%s' is not running", workerID)
		}
		a.Session.PaneSend(cfg.PaneID, "[Main Controller Reply] "+answer)
		fmt.Printf("✓ Replied to worker '%s'\n", workerID)
		return nil
	}

	// v1 fallback: try old role config
	oldConfigPath := internal.ConfigPath(root, a.WtBase, workerID)
	oldCfg, err := internal.LoadRoleConfig(oldConfigPath)
	if err != nil {
		return fmt.Errorf("worker '%s' not found", workerID)
	}

	if !a.Session.PaneAlive(oldCfg.PaneID) {
		return fmt.Errorf("worker '%s' is not running", workerID)
	}

	a.Session.PaneSend(oldCfg.PaneID, "[Main Controller Reply] "+answer)
	fmt.Printf("✓ Replied to '%s'\n", workerID)
	return nil
}
