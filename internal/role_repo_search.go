package internal

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

func (c *RoleRepoGitHubClient) SearchRoleRepos(ctx context.Context, query string) ([]RoleRepoSearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	q := url.Values{}
	q.Set("q", query+" filename:role.yaml in:path")
	q.Set("per_page", "50")

	var payload struct {
		Items []struct {
			Path       string `json:"path"`
			HTMLURL    string `json:"html_url"`
			Repository struct {
				FullName string `json:"full_name"`
				HTMLURL  string `json:"html_url"`
			} `json:"repository"`
		} `json:"items"`
	}

	if err := c.doJSON(ctx, http.MethodGet, "/search/code?"+q.Encode(), &payload); err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	results := make([]RoleRepoSearchResult, 0)
	for _, item := range payload.Items {
		roleName, rolePath, ok := ParseRolePathFromYAMLPath(item.Path)
		if !ok {
			continue
		}
		key := item.Repository.FullName + "::" + rolePath
		if seen[key] {
			continue
		}
		seen[key] = true
		results = append(results, RoleRepoSearchResult{
			Name:      roleName,
			Repo:      item.Repository.FullName,
			RolePath:  rolePath,
			YAMLPath:  item.Path,
			HTMLURL:   item.HTMLURL,
			SourceURL: item.Repository.HTMLURL,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Repo == results[j].Repo {
			return results[i].Name < results[j].Name
		}
		return results[i].Repo < results[j].Repo
	})
	return results, nil
}
