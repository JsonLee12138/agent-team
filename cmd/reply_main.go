// cmd/reply_main.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newReplyMainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   `reply-main "<message>"`,
		Short: "Send a message to the main controller's session (used by workers)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunReplyMain(args[0])
		},
	}
}

func (a *App) RunReplyMain(message string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Try to detect worker ID from worktree directory name
	workerID := filepath.Base(cwd)

	// Try v2: load worker config from agents/workers/<worker-id>/
	root := a.Git.Root()
	configPath := internal.WorkerConfigPath(root, workerID)
	if wcfg, err := internal.LoadWorkerConfig(configPath); err == nil {
		if wcfg.ControllerPaneID == "" {
			return fmt.Errorf("no controller pane ID stored for worker '%s' — was the session opened with agent-team worker open?", workerID)
		}
		if !a.Session.PaneAlive(wcfg.ControllerPaneID) {
			return fmt.Errorf("main controller (pane %s) is not running", wcfg.ControllerPaneID)
		}
		a.Session.PaneSend(wcfg.ControllerPaneID, fmt.Sprintf("[Worker: %s] %s", workerID, message))
		fmt.Printf("✓ Sent to main controller from worker '%s'\n", workerID)
		return nil
	}

	// v1 fallback: try old role config
	roleName, cfg, err := findRoleConfig(cwd)
	if err != nil {
		return err
	}

	if cfg.ControllerPaneID == "" {
		return fmt.Errorf("no controller pane ID stored for role '%s' — was the session opened with agent-team open?", roleName)
	}

	if !a.Session.PaneAlive(cfg.ControllerPaneID) {
		return fmt.Errorf("main controller (pane %s) is not running", cfg.ControllerPaneID)
	}

	a.Session.PaneSend(cfg.ControllerPaneID, fmt.Sprintf("[Role: %s] %s", roleName, message))
	fmt.Printf("✓ Sent to main controller from '%s'\n", roleName)
	return nil
}

// findRoleConfig scans agents/teams/*/config.yaml in the given directory to find the role config (v1 compat).
func findRoleConfig(dir string) (string, *internal.RoleConfig, error) {
	teamsDir := filepath.Join(dir, "agents", "teams")
	entries, err := os.ReadDir(teamsDir)
	if err != nil {
		return "", nil, fmt.Errorf("not in a worker worktree (no agents/teams/ directory found)")
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		configPath := filepath.Join(teamsDir, e.Name(), "config.yaml")
		cfg, err := internal.LoadRoleConfig(configPath)
		if err != nil {
			continue
		}
		return e.Name(), cfg, nil
	}

	return "", nil, fmt.Errorf("no role config found in %s", teamsDir)
}
