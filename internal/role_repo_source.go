package internal

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
)

var roleRepoFullNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

// ParseRoleRepoSource parses owner/repo and GitHub URL forms into normalized source.
func ParseRoleRepoSource(source string) (RoleRepoSource, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return RoleRepoSource{}, fmt.Errorf("empty source")
	}

	if roleRepoFullNamePattern.MatchString(source) {
		parts := strings.SplitN(source, "/", 2)
		return RoleRepoSource{Original: source, Owner: parts[0], Repo: parts[1]}, nil
	}

	if strings.HasPrefix(source, "git@github.com:") {
		repo := strings.TrimPrefix(source, "git@github.com:")
		repo = strings.TrimSuffix(repo, ".git")
		if roleRepoFullNamePattern.MatchString(repo) {
			parts := strings.SplitN(repo, "/", 2)
			return RoleRepoSource{Original: source, Owner: parts[0], Repo: parts[1]}, nil
		}
	}

	u, err := url.Parse(source)
	if err == nil && (u.Scheme == "https" || u.Scheme == "http") && strings.EqualFold(u.Host, "github.com") {
		p := strings.Trim(u.Path, "/")
		p = strings.TrimSuffix(p, ".git")
		parts := strings.Split(p, "/")
		if len(parts) >= 2 {
			repo := parts[0] + "/" + parts[1]
			if roleRepoFullNamePattern.MatchString(repo) {
				return RoleRepoSource{Original: source, Owner: parts[0], Repo: parts[1]}, nil
			}
		}
	}

	return RoleRepoSource{}, fmt.Errorf("unsupported source %q (use owner/repo or GitHub URL)", source)
}

// ParseRolePathFromYAMLPath validates strict role contracts and returns role metadata.
func ParseRolePathFromYAMLPath(p string) (roleName string, rolePath string, ok bool) {
	p = path.Clean(strings.TrimSpace(strings.TrimPrefix(p, "/")))
	parts := strings.Split(p, "/")

	if len(parts) == 4 && parts[0] == "skills" && parts[2] == "references" && parts[3] == "role.yaml" {
		if parts[1] == "" {
			return "", "", false
		}
		return parts[1], strings.Join(parts[:2], "/"), true
	}

	if len(parts) == 5 && parts[0] == ".agent-team" && parts[1] == "teams" && parts[3] == "references" && parts[4] == "role.yaml" {
		if parts[2] == "" {
			return "", "", false
		}
		return parts[2], strings.Join(parts[:3], "/"), true
	}

	if len(parts) == 5 && parts[0] == ".agents" && parts[1] == "teams" && parts[3] == "references" && parts[4] == "role.yaml" {
		if parts[2] == "" {
			return "", "", false
		}
		return parts[2], strings.Join(parts[:3], "/"), true
	}

	return "", "", false
}
