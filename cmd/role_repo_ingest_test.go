package cmd

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
)

type captureReporter struct {
	calls   int
	query   string
	results []internal.RoleRepoSearchResult
	traceID string
}

func (c *captureReporter) ReportAsync(query string, results []internal.RoleRepoSearchResult, traceID string) *sync.WaitGroup {
	c.calls++
	c.query = query
	c.results = append([]internal.RoleRepoSearchResult(nil), results...)
	c.traceID = traceID
	var wg sync.WaitGroup
	return &wg
}

func TestReportRoleRepoInstallIngestReportsInstalledRoles(t *testing.T) {
	origReporter := newRoleHubReporter
	t.Cleanup(func() { newRoleHubReporter = origReporter })

	reporter := &captureReporter{}
	newRoleHubReporter = func() roleHubReporter { return reporter }

	source := internal.RoleRepoSource{Owner: "acme", Repo: "roles"}
	installed := []internal.RoleRepoRemoteRole{
		{Candidate: internal.RoleRepoCandidate{Name: "frontend", RolePath: "skills/frontend", SourceURL: source.HTTPSURL()}},
	}

	reportRoleRepoInstallIngest(source, installed)

	if reporter.calls != 1 {
		t.Fatalf("expected 1 ingest call, got %d", reporter.calls)
	}
	if reporter.query != "install:acme/roles" {
		t.Fatalf("query=%q, want install:acme/roles", reporter.query)
	}
	if len(reporter.results) != 1 {
		t.Fatalf("results len=%d, want 1", len(reporter.results))
	}
	if reporter.results[0].Name != "frontend" || reporter.results[0].RolePath != "skills/frontend" {
		t.Fatalf("unexpected result payload: %+v", reporter.results[0])
	}
	if reporter.traceID == "" {
		t.Fatal("traceID should not be empty")
	}
}

func TestReportRoleRepoInstallIngestSkipsWhenNoInstalledRoles(t *testing.T) {
	origReporter := newRoleHubReporter
	t.Cleanup(func() { newRoleHubReporter = origReporter })

	reporter := &captureReporter{}
	newRoleHubReporter = func() roleHubReporter { return reporter }

	source := internal.RoleRepoSource{Owner: "acme", Repo: "roles"}
	reportRoleRepoInstallIngest(source, nil)

	if reporter.calls != 0 {
		t.Fatalf("expected 0 ingest calls, got %d", reporter.calls)
	}
}

type delayedReporter struct {
	delay time.Duration
}

func (d *delayedReporter) ReportAsync(string, []internal.RoleRepoSearchResult, string) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		time.Sleep(d.delay)
		wg.Done()
	}()
	return &wg
}

func TestReportRoleRepoInstallIngestIsAsyncByDefault(t *testing.T) {
	origDebug := os.Getenv(roleHubIngestDebugEnv)
	t.Cleanup(func() {
		if origDebug == "" {
			os.Unsetenv(roleHubIngestDebugEnv)
			return
		}
		os.Setenv(roleHubIngestDebugEnv, origDebug)
	})
	os.Unsetenv(roleHubIngestDebugEnv)

	origReporter := newRoleHubReporter
	t.Cleanup(func() { newRoleHubReporter = origReporter })

	newRoleHubReporter = func() roleHubReporter { return &delayedReporter{delay: 120 * time.Millisecond} }

	source := internal.RoleRepoSource{Owner: "acme", Repo: "roles"}
	installed := []internal.RoleRepoRemoteRole{
		{Candidate: internal.RoleRepoCandidate{Name: "frontend", RolePath: "skills/frontend", SourceURL: source.HTTPSURL()}},
	}

	start := time.Now()
	reportRoleRepoInstallIngest(source, installed)
	if got := time.Since(start); got > 60*time.Millisecond {
		t.Fatalf("reportRoleRepoInstallIngest should return quickly in async mode, got %v", got)
	}
}

func TestReportRoleRepoInstallIngestWaitsInDebugMode(t *testing.T) {
	origDebug := os.Getenv(roleHubIngestDebugEnv)
	t.Cleanup(func() {
		if origDebug == "" {
			os.Unsetenv(roleHubIngestDebugEnv)
			return
		}
		os.Setenv(roleHubIngestDebugEnv, origDebug)
	})
	os.Setenv(roleHubIngestDebugEnv, "1")

	origReporter := newRoleHubReporter
	t.Cleanup(func() { newRoleHubReporter = origReporter })

	newRoleHubReporter = func() roleHubReporter { return &delayedReporter{delay: 120 * time.Millisecond} }

	source := internal.RoleRepoSource{Owner: "acme", Repo: "roles"}
	installed := []internal.RoleRepoRemoteRole{
		{Candidate: internal.RoleRepoCandidate{Name: "frontend", RolePath: "skills/frontend", SourceURL: source.HTTPSURL()}},
	}

	start := time.Now()
	reportRoleRepoInstallIngest(source, installed)
	if got := time.Since(start); got < 100*time.Millisecond {
		t.Fatalf("debug mode should wait for report completion, got %v", got)
	}
}
