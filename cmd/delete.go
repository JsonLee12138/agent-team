// cmd/delete.go
package cmd

import (
	"fmt"
	"os"

	"github.com/leeforge/agent-team/internal"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Remove a role and its worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunDelete(args[0])
		},
	}
}

func (a *App) RunDelete(name string) error {
	root := a.Git.Root()
	wtPath := internal.WtPath(root, a.WtBase, name)
	configPath := internal.ConfigPath(root, a.WtBase, name)

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("role '%s' not found", name)
	}

	fmt.Printf("Deleting role '%s'...\n", name)

	// Kill running pane if any
	cfg, err := internal.LoadRoleConfig(configPath)
	if err == nil && a.Session.PaneAlive(cfg.PaneID) {
		if killErr := a.Session.KillPane(cfg.PaneID); killErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close pane %s; continuing delete\n", cfg.PaneID)
		}
	}

	// Remove worktree
	if err := a.Git.WorktreeRemove(wtPath); err != nil {
		// Fallback: force remove directory
		os.RemoveAll(wtPath)
	}

	a.Git.DeleteBranch("team/" + name)
	fmt.Printf("✓ Deleted role '%s'\n", name)
	return nil
}
