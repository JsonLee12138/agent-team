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
	roles := internal.ListAvailableRoles(root)
	if len(roles) == 0 {
		fmt.Println("No roles found. Create one using the role-creator skill.")
		return nil
	}

	fmt.Printf("%-24s %s\n", "Role", "Path")
	fmt.Printf("%-24s %s\n", "────────────────────────", "──────────────────────────")
	for _, role := range roles {
		fmt.Printf("%-24s %s\n", role, internal.RoleDir(root, role))
	}
	return nil
}
