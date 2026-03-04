package internal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNormalizeWorkerVerified(t *testing.T) {
	// Set up a mock GitHub API server
	roleYAML := "name: frontend\ndescription: Frontend development role\n"
	encodedYAML := base64.StdEncoding.EncodeToString([]byte(roleYAML))

	mux := http.NewServeMux()
	// GET /repos/acme/roles → default branch
	mux.HandleFunc("/repos/acme/roles", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"default_branch": "main"})
	})
	// GET /repos/acme/roles/git/trees/main → tree with role.yaml
	mux.HandleFunc("/repos/acme/roles/git/trees/main", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"tree": []map[string]string{
				{"path": "skills/frontend/references/role.yaml", "type": "blob", "sha": "sha-role-yaml"},
			},
		})
	})
	// GET /repos/acme/roles/git/blobs/sha-role-yaml → role.yaml content
	mux.HandleFunc("/repos/acme/roles/git/blobs/sha-role-yaml", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"content":  encodedYAML,
			"encoding": "base64",
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewRoleRepoGitHubClientForTest(server.URL, server.Client(), "test-token")
	fixedTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	worker := &NormalizeWorker{
		Client: client,
		NowFn:  func() time.Time { return fixedTime },
	}

	catalog := &RoleRepoCatalog{
		Version: RoleRepoCatalogVersion,
		Entries: []RoleRepoCatalogEntry{
			{
				Name:     "frontend",
				Source:   "acme/roles",
				RolePath: "skills/frontend",
				Status:   CatalogStatusDiscovered,
			},
		},
	}

	results := worker.NormalizeAll(context.Background(), catalog)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != CatalogStatusVerified {
		t.Fatalf("expected verified, got %s: %s", results[0].Status, results[0].Reason)
	}
	if catalog.Entries[0].Status != CatalogStatusVerified {
		t.Fatalf("catalog entry not updated to verified")
	}
	if catalog.Entries[0].VerifiedAt == nil {
		t.Fatal("verifiedAt not set")
	}
}

func TestNormalizeWorkerInvalidSource(t *testing.T) {
	client := NewRoleRepoGitHubClientForTest("http://unused", nil, "")
	fixedTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	worker := &NormalizeWorker{
		Client: client,
		NowFn:  func() time.Time { return fixedTime },
	}

	catalog := &RoleRepoCatalog{
		Version: RoleRepoCatalogVersion,
		Entries: []RoleRepoCatalogEntry{
			{
				Name:     "bad-role",
				Source:   "not-a-valid-source",
				RolePath: "skills/bad-role",
				Status:   CatalogStatusDiscovered,
			},
		},
	}

	results := worker.NormalizeAll(context.Background(), catalog)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != CatalogStatusInvalid {
		t.Fatalf("expected invalid, got %s: %s", results[0].Status, results[0].Reason)
	}
}

func TestNormalizeWorkerUnreachable(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/roles", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"message":"Not Found"}`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewRoleRepoGitHubClientForTest(server.URL, server.Client(), "test-token")
	fixedTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	worker := &NormalizeWorker{
		Client: client,
		NowFn:  func() time.Time { return fixedTime },
	}

	catalog := &RoleRepoCatalog{
		Version: RoleRepoCatalogVersion,
		Entries: []RoleRepoCatalogEntry{
			{
				Name:     "frontend",
				Source:   "acme/roles",
				RolePath: "skills/frontend",
				Status:   CatalogStatusDiscovered,
			},
		},
	}

	results := worker.NormalizeAll(context.Background(), catalog)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != CatalogStatusUnreachable {
		t.Fatalf("expected unreachable, got %s: %s", results[0].Status, results[0].Reason)
	}
}

func TestNormalizeWorkerInvalidYAML(t *testing.T) {
	// role.yaml with missing name
	roleYAML := "description: some role\n"
	encodedYAML := base64.StdEncoding.EncodeToString([]byte(roleYAML))

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/roles", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"default_branch": "main"})
	})
	mux.HandleFunc("/repos/acme/roles/git/trees/main", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"tree": []map[string]string{
				{"path": "skills/frontend/references/role.yaml", "type": "blob", "sha": "sha-role-yaml"},
			},
		})
	})
	mux.HandleFunc("/repos/acme/roles/git/blobs/sha-role-yaml", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"content":  encodedYAML,
			"encoding": "base64",
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewRoleRepoGitHubClientForTest(server.URL, server.Client(), "test-token")
	worker := &NormalizeWorker{
		Client: client,
		NowFn:  func() time.Time { return time.Now() },
	}

	catalog := &RoleRepoCatalog{
		Version: RoleRepoCatalogVersion,
		Entries: []RoleRepoCatalogEntry{
			{
				Name:     "frontend",
				Source:   "acme/roles",
				RolePath: "skills/frontend",
				Status:   CatalogStatusDiscovered,
			},
		},
	}

	results := worker.NormalizeAll(context.Background(), catalog)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != CatalogStatusInvalid {
		t.Fatalf("expected invalid for missing name, got %s: %s", results[0].Status, results[0].Reason)
	}
}

func TestNormalizeWorkerSkipsNonDiscovered(t *testing.T) {
	client := NewRoleRepoGitHubClientForTest("http://unused", nil, "")
	worker := &NormalizeWorker{
		Client: client,
		NowFn:  time.Now,
	}

	catalog := &RoleRepoCatalog{
		Version: RoleRepoCatalogVersion,
		Entries: []RoleRepoCatalogEntry{
			{Name: "verified-role", Status: CatalogStatusVerified},
			{Name: "invalid-role", Status: CatalogStatusInvalid},
		},
	}

	results := worker.NormalizeAll(context.Background(), catalog)
	if len(results) != 0 {
		t.Fatalf("expected 0 results (nothing discovered), got %d", len(results))
	}
}

func TestNormalizeWorkerRoleNotInTree(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/roles", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"default_branch": "main"})
	})
	mux.HandleFunc("/repos/acme/roles/git/trees/main", func(w http.ResponseWriter, r *http.Request) {
		// Empty tree - no roles
		json.NewEncoder(w).Encode(map[string]any{
			"tree": []map[string]string{},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewRoleRepoGitHubClientForTest(server.URL, server.Client(), "test-token")
	worker := &NormalizeWorker{
		Client: client,
		NowFn:  time.Now,
	}

	catalog := &RoleRepoCatalog{
		Version: RoleRepoCatalogVersion,
		Entries: []RoleRepoCatalogEntry{
			{
				Name:     "frontend",
				Source:   "acme/roles",
				RolePath: "skills/frontend",
				Status:   CatalogStatusDiscovered,
			},
		},
	}

	results := worker.NormalizeAll(context.Background(), catalog)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != CatalogStatusInvalid {
		t.Fatalf("expected invalid (role not found), got %s: %s", results[0].Status, results[0].Reason)
	}
}

func TestCatalogFromDiscoveredRoles(t *testing.T) {
	now := time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC)
	roles := []RoleRepoRemoteRole{
		{
			Candidate: RoleRepoCandidate{
				Name:       "frontend",
				RolePath:   "skills/frontend",
				Source:     RoleRepoSource{Owner: "acme", Repo: "roles"},
				SourceType: "github",
				SourceURL:  "https://github.com/acme/roles",
			},
			FolderHash: "hash123",
		},
	}

	entries := CatalogFromDiscoveredRoles(roles, func() time.Time { return now })
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Status != CatalogStatusDiscovered {
		t.Fatalf("expected discovered status, got %s", entries[0].Status)
	}
	if entries[0].Source != "acme/roles" {
		t.Fatalf("expected source acme/roles, got %s", entries[0].Source)
	}
}

func TestFormatNormalizeResults(t *testing.T) {
	results := []NormalizeResult{
		{
			Entry:  RoleRepoCatalogEntry{Name: "a", Source: "acme/roles"},
			Status: CatalogStatusVerified,
			Reason: "all checks passed",
		},
		{
			Entry:  RoleRepoCatalogEntry{Name: "b", Source: "acme/roles"},
			Status: CatalogStatusInvalid,
			Reason: "missing name",
		},
	}

	output := FormatNormalizeResults(results)
	if output == "" {
		t.Fatal("expected non-empty output")
	}
	if !containsInsensitive(output, "verified") {
		t.Fatal("expected 'verified' in output")
	}
	if !containsInsensitive(output, "invalid") {
		t.Fatal("expected 'invalid' in output")
	}

	// Empty results
	empty := FormatNormalizeResults(nil)
	if empty == "" {
		t.Fatal("expected non-empty output for nil results")
	}
}
