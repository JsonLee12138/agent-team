package internal

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRoleRepoPaths_ProjectScope(t *testing.T) {
	root := "/tmp/project"
	install, err := ResolveRoleRepoInstallRoot(root, RoleRepoScopeProject)
	if err != nil {
		t.Fatalf("ResolveRoleRepoInstallRoot: %v", err)
	}
	if !strings.HasSuffix(install, filepath.Join(".agents", "teams")) {
		t.Fatalf("project install root = %q", install)
	}

	lockPath, err := ResolveRoleRepoLockPath(root, RoleRepoScopeProject)
	if err != nil {
		t.Fatalf("ResolveRoleRepoLockPath: %v", err)
	}
	want := filepath.Join(root, "roles-lock.json")
	if lockPath != want {
		t.Fatalf("lock path = %q, want %q", lockPath, want)
	}
}

func TestResolveRoleRepoScope(t *testing.T) {
	if got := ResolveRoleRepoScope(false); got != RoleRepoScopeProject {
		t.Fatalf("scope false = %s", got)
	}
	if got := ResolveRoleRepoScope(true); got != RoleRepoScopeGlobal {
		t.Fatalf("scope true = %s", got)
	}
}
