// cmd/reply.go
package cmd

import (
	"fmt"

	"github.com/leeforge/agent-team/internal"
	"github.com/spf13/cobra"
)

func newReplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   `reply <name> "<answer>"`,
		Short: "Send a reply to a role's running session",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunReply(args[0], args[1])
		},
	}
}

func (a *App) RunReply(name, answer string) error {
	root := a.Git.Root()
	configPath := internal.ConfigPath(root, a.WtBase, name)

	cfg, err := internal.LoadRoleConfig(configPath)
	if err != nil {
		return fmt.Errorf("role '%s' not found", name)
	}

	if !a.Session.PaneAlive(cfg.PaneID) {
		return fmt.Errorf("role '%s' is not running", name)
	}

	a.Session.PaneSend(cfg.PaneID, "[Main Controller Reply] "+answer)
	fmt.Printf("✓ Replied to '%s'\n", name)
	return nil
}
