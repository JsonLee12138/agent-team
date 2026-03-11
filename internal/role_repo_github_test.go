package internal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type mockSearchItem struct {
	Path     string
	HTMLURL  string
	Repo     string
	RepoHTML string
}

func newRoleRepoMockGitHubServer(t *testing.T, defaultBranch string, tree []roleRepoTreeEntry, blobs map[string]string, searchItems []mockSearchItem) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/repos/acme/roles":
			_ = json.NewEncoder(w).Encode(map[string]any{"default_branch": defaultBranch})
		case strings.HasPrefix(r.URL.Path, "/repos/acme/roles/git/trees/"):
			_ = json.NewEncoder(w).Encode(map[string]any{"tree": tree})
		case strings.HasPrefix(r.URL.Path, "/repos/acme/roles/git/blobs/"):
			sha := strings.TrimPrefix(r.URL.Path, "/repos/acme/roles/git/blobs/")
			content, ok := blobs[sha]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"message": "not found"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"encoding": "base64",
				"content":  base64.StdEncoding.EncodeToString([]byte(content)),
			})
		case r.URL.Path == "/search/code":
			items := make([]map[string]any, 0, len(searchItems))
			for _, item := range searchItems {
				items = append(items, map[string]any{
					"path":     item.Path,
					"html_url": item.HTMLURL,
					"repository": map[string]any{
						"full_name": item.Repo,
						"html_url":  item.RepoHTML,
					},
				})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"message": "unsupported path"})
		}
	}))
}

func TestRoleRepoSearchFiltersAndDeduplicates(t *testing.T) {
	srv := newRoleRepoMockGitHubServer(t, "main", nil, nil, []mockSearchItem{
		{Path: "skills/frontend/references/role.yaml", HTMLURL: "https://example.com/1", Repo: "acme/roles", RepoHTML: "https://github.com/acme/roles"},
		{Path: "skills/frontend/references/role.yaml", HTMLURL: "https://example.com/dup", Repo: "acme/roles", RepoHTML: "https://github.com/acme/roles"},
		{Path: "agents/teams/bad/references/role.yaml", HTMLURL: "https://example.com/2", Repo: "acme/roles", RepoHTML: "https://github.com/acme/roles"},
	})
	defer srv.Close()

	client := NewRoleRepoGitHubClientForTest(srv.URL, srv.Client(), "")
	results, err := client.SearchRoleRepos(context.Background(), "frontend")
	if err != nil {
		t.Fatalf("SearchRoleRepos: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results len = %d, want 1", len(results))
	}
	if results[0].Name != "frontend" || results[0].Repo != "acme/roles" {
		t.Fatalf("unexpected result: %+v", results[0])
	}
}

func TestRoleRepoDiscoverRemoteRolesStrictPathContracts(t *testing.T) {
	tree := []roleRepoTreeEntry{
		{Path: "skills/frontend/references/role.yaml", Type: "blob", SHA: "sha-role-1"},
		{Path: "skills/frontend/SKILL.md", Type: "blob", SHA: "sha-skill-1"},
		{Path: ".agents/teams/backend/references/role.yaml", Type: "blob", SHA: "sha-role-2"},
		{Path: ".agents/teams/backend/system.md", Type: "blob", SHA: "sha-system-2"},
		{Path: "agents/teams/legacy/references/role.yaml", Type: "blob", SHA: "sha-bad"},
	}
	srv := newRoleRepoMockGitHubServer(t, "main", tree, nil, nil)
	defer srv.Close()

	client := NewRoleRepoGitHubClientForTest(srv.URL, srv.Client(), "")
	source, _ := ParseRoleRepoSource("acme/roles")
	roles, err := client.DiscoverRemoteRoles(context.Background(), source)
	if err != nil {
		t.Fatalf("DiscoverRemoteRoles: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("roles len = %d, want 2", len(roles))
	}
	if roles[0].Candidate.Name != "backend" || roles[1].Candidate.Name != "frontend" {
		t.Fatalf("unexpected roles: %+v", roles)
	}
	for _, role := range roles {
		if role.FolderHash == "" {
			t.Fatalf("empty folder hash for role %s", role.Candidate.Name)
		}
	}
}

func TestInstallRoleRepoRemoteRoleAndCheckUpdate(t *testing.T) {
	treeV1 := []roleRepoTreeEntry{
		{Path: "skills/frontend/references/role.yaml", Type: "blob", SHA: "sha-role-v1"},
		{Path: "skills/frontend/SKILL.md", Type: "blob", SHA: "sha-skill-v1"},
	}
	blobsV1 := map[string]string{
		"sha-role-v1":  "name: frontend\n",
		"sha-skill-v1": "# frontend\n",
	}

	// Dynamic tree/blob state to simulate remote update.
	currentTree := treeV1
	currentBlobs := blobsV1
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/repos/acme/roles":
			_ = json.NewEncoder(w).Encode(map[string]any{"default_branch": "main"})
		case strings.HasPrefix(r.URL.Path, "/repos/acme/roles/git/trees/"):
			_ = json.NewEncoder(w).Encode(map[string]any{"tree": currentTree})
		case strings.HasPrefix(r.URL.Path, "/repos/acme/roles/git/blobs/"):
			sha := strings.TrimPrefix(r.URL.Path, "/repos/acme/roles/git/blobs/")
			content := currentBlobs[sha]
			_ = json.NewEncoder(w).Encode(map[string]any{"encoding": "base64", "content": base64.StdEncoding.EncodeToString([]byte(content))})
		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"message": "unsupported path"})
		}
	}))
	defer srv.Close()

	client := NewRoleRepoGitHubClientForTest(srv.URL, srv.Client(), "")
	source, _ := ParseRoleRepoSource("acme/roles")
	roles, err := client.DiscoverRemoteRoles(context.Background(), source)
	if err != nil {
		t.Fatalf("DiscoverRemoteRoles: %v", err)
	}
	if len(roles) != 1 {
		t.Fatalf("roles len = %d, want 1", len(roles))
	}

	installRoot := filepath.Join(t.TempDir(), ".agents", "teams")
	now := time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC)
	nowFn := func() time.Time { return now }

	targetDir, err := InstallRoleRepoRemoteRole(context.Background(), client, roles[0], installRoot, false)
	if err != nil {
		t.Fatalf("InstallRoleRepoRemoteRole: %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetDir, "references", "role.yaml")); err != nil {
		t.Fatalf("expected installed role.yaml: %v", err)
	}

	if _, err := InstallRoleRepoRemoteRole(context.Background(), client, roles[0], installRoot, false); err == nil {
		t.Fatal("expected conflict error on reinstall without overwrite")
	}

	lock := RoleRepoLockFile{Version: 1, Entries: []RoleRepoLockEntry{{
		Name:        "frontend",
		Source:      "acme/roles",
		SourceType:  "github",
		SourceURL:   "https://github.com/acme/roles",
		RolePath:    "skills/frontend",
		FolderHash:  roles[0].FolderHash,
		InstalledAt: now,
		UpdatedAt:   now,
	}}}

	localOnly := filepath.Join(installRoot, "local-only")
	if err := os.MkdirAll(localOnly, 0755); err != nil {
		t.Fatal(err)
	}

	// mutate remote role
	currentTree = []roleRepoTreeEntry{
		{Path: "skills/frontend/references/role.yaml", Type: "blob", SHA: "sha-role-v2"},
		{Path: "skills/frontend/SKILL.md", Type: "blob", SHA: "sha-skill-v2"},
	}
	currentBlobs = map[string]string{
		"sha-role-v2":  "name: frontend\nversion: 2\n",
		"sha-skill-v2": "# frontend v2\n",
	}

	statuses, untracked := CheckRoleRepoUpdates(context.Background(), client, installRoot, lock)
	if len(statuses) != 1 || statuses[0].State != "update_available" {
		t.Fatalf("unexpected check statuses: %+v", statuses)
	}
	if len(untracked) != 1 || untracked[0] != "local-only" {
		t.Fatalf("unexpected untracked: %+v", untracked)
	}

	updated, skipped, failed := UpdateRoleRepoFromLock(context.Background(), client, installRoot, &lock, false, nil, nowFn)
	if len(updated) != 0 || len(skipped) != 1 || len(failed) != 0 {
		t.Fatalf("expected skip without overwrite, got updated=%v skipped=%v failed=%v", updated, skipped, failed)
	}

	updated, skipped, failed = UpdateRoleRepoFromLock(context.Background(), client, installRoot, &lock, true, nil, nowFn)
	if len(updated) != 1 || updated[0] != "frontend" || len(failed) != 0 {
		t.Fatalf("unexpected update result: updated=%v skipped=%v failed=%v", updated, skipped, failed)
	}
	if lock.Entries[0].FolderHash == roles[0].FolderHash {
		t.Fatalf("expected updated folder hash, got unchanged %s", lock.Entries[0].FolderHash)
	}

	data, err := os.ReadFile(filepath.Join(installRoot, "frontend", "references", "role.yaml"))
	if err != nil {
		t.Fatalf("read updated role.yaml: %v", err)
	}
	if !strings.Contains(string(data), "version: 2") {
		t.Fatalf("expected updated content, got %q", string(data))
	}
}
