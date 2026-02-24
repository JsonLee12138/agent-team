// cmd/status.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show all roles, running state, and pending task count",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunStatus()
		},
	}
}

func (a *App) RunStatus() error {
	root := a.Git.Root()
	roles := internal.ListRoles(root, a.WtBase)
	if len(roles) == 0 {
		fmt.Println("No roles found. Create one with: agent-team create <name>")
		return nil
	}

	fmt.Printf("%-16s %-24s %s\n", "Role", "Status", "Pending Tasks")
	fmt.Printf("%-16s %-24s %s\n", "────────────────", "────────────────────────", "─────────────")

	for _, role := range roles {
		configPath := internal.ConfigPath(root, a.WtBase, role)
		pendingDir := filepath.Join(internal.TeamsDir(root, a.WtBase, role), "tasks", "pending")

		cfg, _ := internal.LoadRoleConfig(configPath)
		status := "✗ offline"
		if cfg != nil && a.Session.PaneAlive(cfg.PaneID) {
			status = fmt.Sprintf("✓ running [p:%s]", cfg.PaneID)
		}

		count := 0
		if entries, err := os.ReadDir(pendingDir); err == nil {
			for _, e := range entries {
				if filepath.Ext(e.Name()) == ".md" {
					count++
				}
			}
		}

		fmt.Printf("%-16s %-24s %d\n", role, status, count)
	}
	return nil
}
