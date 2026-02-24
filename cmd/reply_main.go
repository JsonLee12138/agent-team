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
		Short: "Send a message to the main controller's session (used by roles)",
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

	// Find role config in current worktree: agents/teams/*/config.yaml
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

// findRoleConfig scans agents/teams/*/config.yaml in the given directory to find the role config.
func findRoleConfig(dir string) (string, *internal.RoleConfig, error) {
	teamsDir := filepath.Join(dir, "agents", "teams")
	entries, err := os.ReadDir(teamsDir)
	if err != nil {
		return "", nil, fmt.Errorf("not in a role worktree (no agents/teams/ directory found)")
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
