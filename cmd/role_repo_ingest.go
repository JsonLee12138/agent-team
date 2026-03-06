package cmd

import (
	"context"
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

	wg := newRoleHubReporter().ReportAsync("install:"+source.FullName(), results, internal.GenerateTraceID())
	if roleHubIngestDebugEnabled() && wg != nil {
		wg.Wait()
	}
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
