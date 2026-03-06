package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/JsonLee12138/agent-team/internal"
)

const roleHubIngestDebugEnv = "AGENT_TEAM_ROLE_HUB_DEBUG"

type roleRepoSearchClient interface {
	SearchRoleRepos(ctx context.Context, query string) ([]internal.RoleRepoSearchResult, error)
}

type roleHubReporter interface {
	ReportAsync(query string, results []internal.RoleRepoSearchResult, traceID string) *sync.WaitGroup
}

var newRoleRepoSearchClient = func() roleRepoSearchClient {
	return internal.NewRoleRepoGitHubClient()
}

var newRoleHubReporter = func() roleHubReporter {
	return internal.NewIngestClient()
}

func reportRoleRepoInstallIngest(source internal.RoleRepoSource, installed []internal.RoleRepoRemoteRole) {
	if len(installed) == 0 {
		return
	}

	results := make([]internal.RoleRepoSearchResult, 0, len(installed))
	for _, role := range installed {
		sourceURL := strings.TrimSpace(role.Candidate.SourceURL)
		if sourceURL == "" {
			sourceURL = source.HTTPSURL()
		}
		results = append(results, internal.RoleRepoSearchResult{
			Name:      role.Candidate.Name,
			Repo:      source.FullName(),
			RolePath:  role.Candidate.RolePath,
			YAMLPath:  role.Candidate.YAMLPath,
			SourceURL: sourceURL,
		})
	}

	traceID := internal.GenerateTraceID()
	wg := newRoleHubReporter().ReportAsync("install:"+source.FullName(), results, traceID)
	if wg == nil {
		return
	}

	if roleHubIngestDebugEnabled() {
		fmt.Fprintf(os.Stderr, "[role-hub-ingest] waiting for install report trace=%s roles=%d\n", traceID, len(results))
		wg.Wait()
		fmt.Fprintf(os.Stderr, "[role-hub-ingest] install report finished trace=%s\n", traceID)
		return
	}
	wg.Wait()
}

func roleHubIngestDebugEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(roleHubIngestDebugEnv)))
	switch v {
	case "1", "true", "yes", "on", "debug":
		return true
	default:
		return false
	}
}
