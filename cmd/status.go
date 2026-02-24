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
		Short: "Show all roles, running state, and OpenSpec change status",
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

	fmt.Printf("%-16s %-24s %s\n", "Role", "Status", "Changes")
	fmt.Printf("%-16s %-24s %s\n", "────────────────", "────────────────────────", "──────────────────────────")

	for _, role := range roles {
		configPath := internal.ConfigPath(root, a.WtBase, role)
		wtPath := internal.WtPath(root, a.WtBase, role)

		cfg, _ := internal.LoadRoleConfig(configPath)
		status := "✗ offline"
		if cfg != nil && a.Session.PaneAlive(cfg.PaneID) {
			status = fmt.Sprintf("✓ running [p:%s]", cfg.PaneID)
		}

		// Count changes by reading openspec/changes/ directory
		changesDir := filepath.Join(wtPath, "openspec", "changes")
		changesSummary := "0"
		if entries, err := os.ReadDir(changesDir); err == nil {
			active := 0
			for _, e := range entries {
				if e.IsDir() && e.Name() != "archive" {
					active++
				}
			}
			if active > 0 {
				changesSummary = fmt.Sprintf("%d active", active)
			} else {
				changesSummary = "0"
			}
		}

		fmt.Printf("%-16s %-24s %s\n", role, status, changesSummary)
	}
	return nil
}
