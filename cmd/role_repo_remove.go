package cmd

import (
	"fmt"
	"io"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newRoleRepoRemoveCmd() *cobra.Command {
	var global bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "remove [roles...]",
		Short: "Remove installed repository-managed roles",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunRoleRepoRemove(cmd.InOrStdin(), cmd.OutOrStdout(), args, global, yes)
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Remove from global scope")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Confirm destructive removal")
	return cmd
}

func (a *App) RunRoleRepoRemove(in io.Reader, out io.Writer, roleNames []string, global bool, yes bool) error {
	root := a.Git.Root()
	scope := roleRepoScopeFromFlag(global)
	installRoot, err := internal.ResolveRoleRepoInstallRoot(root, scope)
	if err != nil {
		return err
	}
	lockPath, lock, warning, err := roleRepoLockForScope(root, scope)
	if err != nil {
		return err
	}
	printRoleRepoLockWarning(warning)

	if len(roleNames) == 0 {
		installedNames, listErr := internal.ListInstalledRoleRepoNames(installRoot)
		if listErr != nil {
			return listErr
		}
		roleNames = installedNames
		if len(roleNames) > 1 {
			chosen, selectErr := promptSelectNames(in, out, "Select role(s) to remove:", roleNames)
			if selectErr != nil {
				return selectErr
			}
			roleNames = chosen
		}
		if !yes {
			ok, confirmErr := promptConfirm(in, out, fmt.Sprintf("Remove %d role(s)?", len(roleNames)))
			if confirmErr != nil {
				return confirmErr
			}
			if !ok {
				return fmt.Errorf("remove cancelled")
			}
		}
	} else if !yes {
		ok, confirmErr := promptConfirm(in, out, fmt.Sprintf("Remove %d role(s)?", len(roleNames)))
		if confirmErr != nil {
			return confirmErr
		}
		if !ok {
			return fmt.Errorf("remove cancelled")
		}
	}
	if len(roleNames) == 0 {
		fmt.Fprintln(out, "No roles to remove.")
		return nil
	}

	removed, missing, failed := internal.RemoveInstalledRoleRepo(installRoot, roleNames)
	if len(removed) > 0 {
		internal.RemoveRoleRepoLockEntries(&lock, removed)
		if err := internal.WriteRoleRepoLock(lockPath, lock); err != nil {
			return err
		}
	}

	for _, name := range removed {
		fmt.Fprintf(out, "- removed %s\n", name)
	}
	for _, name := range missing {
		fmt.Fprintf(out, "- skipped %s (not installed)\n", name)
	}
	for name, rmErr := range failed {
		fmt.Fprintf(out, "- failed %s: %v\n", name, rmErr)
	}
	if len(failed) > 0 {
		return fmt.Errorf("one or more removals failed")
	}
	return nil
}
