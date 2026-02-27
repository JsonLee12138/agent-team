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
	configPath := internal.WorkerConfigPath(root, workerID)
	cfg, err := internal.LoadWorkerConfig(configPath)
	if err != nil {
		return fmt.Errorf("worker '%s' not found: %w", workerID, err)
	}

	if !a.Session.PaneAlive(cfg.PaneID) {
		return fmt.Errorf("worker '%s' is not running", workerID)
	}

	a.Session.PaneSend(cfg.PaneID, "[Main Controller Reply] "+answer)
	fmt.Printf("✓ Replied to worker '%s'\n", workerID)
	return nil
}
