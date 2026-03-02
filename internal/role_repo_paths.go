package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

func ResolveRoleRepoScope(global bool) RoleRepoScope {
	if global {
		return RoleRepoScopeGlobal
	}
	return RoleRepoScopeProject
}

func ResolveRoleRepoInstallRoot(root string, scope RoleRepoScope) (string, error) {
	switch scope {
	case RoleRepoScopeProject:
		return filepath.Join(ResolveAgentsDir(root), "teams"), nil
	case RoleRepoScopeGlobal:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		return filepath.Join(home, ".agents", "roles"), nil
	default:
		return "", fmt.Errorf("unknown scope: %s", scope)
	}
}

func ResolveRoleRepoLockPath(root string, scope RoleRepoScope) (string, error) {
	switch scope {
	case RoleRepoScopeProject:
		return filepath.Join(root, "roles-lock.json"), nil
	case RoleRepoScopeGlobal:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		return filepath.Join(home, ".agents", ".role-lock.json"), nil
	default:
		return "", fmt.Errorf("unknown scope: %s", scope)
	}
}
