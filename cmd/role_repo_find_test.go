package cmd

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

type fakeRoleRepoSearchClient struct {
	results []internal.RoleRepoSearchResult
	err     error
}

func (f *fakeRoleRepoSearchClient) SearchRoleRepos(context.Context, string) ([]internal.RoleRepoSearchResult, error) {
	return f.results, f.err
}

type fakeRoleHubReporter struct {
	calls int
}

func (f *fakeRoleHubReporter) ReportAsync(string, []internal.RoleRepoSearchResult, string) *sync.WaitGroup {
	f.calls++
	var wg sync.WaitGroup
	return &wg
}

func TestRoleRepoFindDirectDoesNotIngest(t *testing.T) {
	origSearchClient := newRoleRepoSearchClient
	origReporter := newRoleHubReporter
	t.Cleanup(func() {
		newRoleRepoSearchClient = origSearchClient
		newRoleHubReporter = origReporter
	})

	newRoleRepoSearchClient = func() roleRepoSearchClient {
		return &fakeRoleRepoSearchClient{results: []internal.RoleRepoSearchResult{{
			Name:     "frontend",
			Repo:     "acme/roles",
			RolePath: "skills/frontend",
		}}}
	}

	reporter := &fakeRoleHubReporter{}
	newRoleHubReporter = func() roleHubReporter { return reporter }

	var out bytes.Buffer
	app := &App{}
	if err := app.roleRepoFindDirect(&out, "frontend"); err != nil {
		t.Fatalf("roleRepoFindDirect: %v", err)
	}

	if reporter.calls != 0 {
		t.Fatalf("expected no ingest calls from find, got %d", reporter.calls)
	}
}
