package cmd

import (
	"fmt"
	"sort"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newRoleRepoListCmd() *cobra.Command {
	var global bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed repository-managed roles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunRoleRepoList(global)
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, "List global role installs")
	return cmd
}

func (a *App) RunRoleRepoList(global bool) error {
	root := a.Git.Root()
	scope := roleRepoScopeFromFlag(global)
	installRoot, err := internal.ResolveRoleRepoInstallRoot(root, scope)
	if err != nil {
		return err
	}

	names, err := internal.ListInstalledRoleRepoNames(installRoot)
	if err != nil {
		return err
	}
	lockPath, lock, warning, err := roleRepoLockForScope(root, scope)
	if err != nil {
		return err
	}
	_ = lockPath
	printRoleRepoLockWarning(warning)

	if len(names) == 0 {
		fmt.Println("No installed roles found.")
		return nil
	}

	sort.Strings(names)
	fmt.Printf("%-20s %-32s %s\n", "Role", "Source", "Role Path")
	fmt.Printf("%-20s %-32s %s\n", "────────────────────", "────────────────────────────────", "────────────────────────")
	for _, name := range names {
		entry, ok := internal.FindRoleRepoLockEntry(lock, name)
		if ok {
			fmt.Printf("%-20s %-32s %s\n", name, entry.Source, entry.RolePath)
			continue
		}
		fmt.Printf("%-20s %-32s %s\n", name, "(untracked)", "-")
	}
	return nil
}
