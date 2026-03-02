package internal

import (
	"context"
	"sort"
	"strings"
)

// RoleRepoRemoteRole bundles candidate metadata with tree files/hash for install/update flows.
type RoleRepoRemoteRole struct {
	Candidate  RoleRepoCandidate
	Files      []roleRepoTreeEntry
	FolderHash string
}

func (c *RoleRepoGitHubClient) DiscoverRemoteRoles(ctx context.Context, source RoleRepoSource) ([]RoleRepoRemoteRole, error) {
	defaultBranch, err := c.getDefaultBranch(ctx, source)
	if err != nil {
		return nil, err
	}
	tree, err := c.getRepoTree(ctx, source, defaultBranch)
	if err != nil {
		return nil, err
	}

	roleByPath := map[string]*RoleRepoRemoteRole{}
	for _, e := range tree {
		if e.Type != "blob" {
			continue
		}
		roleName, rolePath, ok := ParseRolePathFromYAMLPath(e.Path)
		if !ok {
			continue
		}
		candidate := RoleRepoCandidate{
			Name:       roleName,
			RolePath:   rolePath,
			YAMLPath:   e.Path,
			Source:     source,
			SourceType: "github",
			SourceURL:  source.HTTPSURL(),
		}
		roleByPath[rolePath] = &RoleRepoRemoteRole{Candidate: candidate}
	}

	if len(roleByPath) == 0 {
		return []RoleRepoRemoteRole{}, nil
	}

	for _, e := range tree {
		if e.Type != "blob" {
			continue
		}
		for rolePath, role := range roleByPath {
			prefix := rolePath + "/"
			if strings.HasPrefix(e.Path, prefix) {
				role.Files = append(role.Files, e)
			}
		}
	}

	roles := make([]RoleRepoRemoteRole, 0, len(roleByPath))
	for _, role := range roleByPath {
		role.FolderHash = hashRoleRepoTreeFiles(role.Files)
		sort.Slice(role.Files, func(i, j int) bool {
			return role.Files[i].Path < role.Files[j].Path
		})
		roles = append(roles, *role)
	}
	sort.Slice(roles, func(i, j int) bool {
		return roles[i].Candidate.Name < roles[j].Candidate.Name
	})
	return roles, nil
}
