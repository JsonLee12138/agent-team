// cmd/role.go
package cmd

import (
	"fmt"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newRoleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "role",
		Short: "Manage roles (skill package definitions)",
	}
	cmd.AddCommand(newRoleListCmd())
	cmd.AddCommand(newRoleCreateCmd())
	return cmd
}

func newRoleListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available roles in agents/teams/",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunRoleList()
		},
	}
}

func (a *App) RunRoleList() error {
	root := a.Git.Root()

	// Project roles
	projectRoles := internal.ListAvailableRoles(root)
	// Global roles
	globalRoles, _ := internal.ListGlobalRoles()

	if len(projectRoles) == 0 && len(globalRoles) == 0 {
		fmt.Println("No roles found. Create one using the role-creator skill.")
		return nil
	}

	if len(projectRoles) > 0 {
		fmt.Println("Project roles (.agents/teams/):")
		fmt.Printf("  %-24s %s\n", "Role", "Path")
		fmt.Printf("  %-24s %s\n", "────────────────────────", "──────────────────────────")
		for _, role := range projectRoles {
			fmt.Printf("  %-24s %s\n", role, internal.RoleDir(root, role))
		}
		fmt.Println()
	}

	if len(globalRoles) > 0 {
		fmt.Println("Global roles (~/.agents/roles/):")
		fmt.Printf("  %-24s %s\n", "Role", "Path")
		fmt.Printf("  %-24s %s\n", "────────────────────────", "──────────────────────────")
		for _, r := range globalRoles {
			fmt.Printf("  %-24s %s\n", r.RoleName, r.Path)
		}
		fmt.Println()
	}

	fmt.Printf("Total: %d role(s) (%d project, %d global)\n", len(projectRoles)+len(globalRoles), len(projectRoles), len(globalRoles))
	return nil
}
