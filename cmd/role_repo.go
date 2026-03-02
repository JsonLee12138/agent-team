package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newRoleRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "role-repo",
		Short: "Manage repository-backed role installations",
	}
	cmd.AddCommand(newRoleRepoSearchCmd())
	cmd.AddCommand(newRoleRepoAddCmd())
	cmd.AddCommand(newRoleRepoListCmd())
	cmd.AddCommand(newRoleRepoRemoveCmd())
	cmd.AddCommand(newRoleRepoCheckCmd())
	cmd.AddCommand(newRoleRepoUpdateCmd())
	return cmd
}

func roleRepoScopeFromFlag(global bool) internal.RoleRepoScope {
	return internal.ResolveRoleRepoScope(global)
}

func roleRepoLockForScope(root string, scope internal.RoleRepoScope) (path string, lock internal.RoleRepoLockFile, warning error, err error) {
	path, err = internal.ResolveRoleRepoLockPath(root, scope)
	if err != nil {
		return "", internal.RoleRepoLockFile{}, nil, err
	}
	lock, err = internal.ReadRoleRepoLock(path)
	if err != nil {
		if errors.Is(err, internal.ErrRoleRepoLockCorrupt) {
			warning = err
			return path, lock, warning, nil
		}
		return "", internal.RoleRepoLockFile{}, nil, err
	}
	return path, lock, nil, nil
}

func printRoleRepoLockWarning(err error) {
	if err == nil {
		return
	}
	fmt.Printf("Warning: %s; recovered with empty lock.\n", strings.TrimSpace(err.Error()))
}
