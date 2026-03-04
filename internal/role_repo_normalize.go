package internal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// NormalizeResult captures the outcome of normalizing one discovered role.
type NormalizeResult struct {
	Entry  RoleRepoCatalogEntry
	Status RoleRepoCatalogStatus
	Reason string
	Err    error
}

// NormalizeWorker validates discovered roles and transitions them to verified/invalid/unreachable.
type NormalizeWorker struct {
	Client *RoleRepoGitHubClient
	NowFn  func() time.Time
}

// NewNormalizeWorker creates a worker with the given GitHub client.
func NewNormalizeWorker(client *RoleRepoGitHubClient) *NormalizeWorker {
	return &NormalizeWorker{
		Client: client,
		NowFn:  time.Now,
	}
}

// NormalizeAll processes all discovered entries in the catalog and updates their status.
func (w *NormalizeWorker) NormalizeAll(ctx context.Context, catalog *RoleRepoCatalog) []NormalizeResult {
	var results []NormalizeResult
	for i := range catalog.Entries {
		entry := &catalog.Entries[i]
		if entry.Status != CatalogStatusDiscovered {
			continue
		}
		result := w.normalize(ctx, entry)
		results = append(results, result)
	}
	return results
}

// NormalizeEntry processes a single entry by source+name.
func (w *NormalizeWorker) NormalizeEntry(ctx context.Context, catalog *RoleRepoCatalog, source, name string) (NormalizeResult, bool) {
	for i := range catalog.Entries {
		entry := &catalog.Entries[i]
		if entry.Source == source && entry.Name == name {
			return w.normalize(ctx, entry), true
		}
	}
	return NormalizeResult{}, false
}

func (w *NormalizeWorker) normalize(ctx context.Context, entry *RoleRepoCatalogEntry) NormalizeResult {
	result := NormalizeResult{Entry: *entry}

	// Step 1: Parse source
	source, err := ParseRoleRepoSource(entry.Source)
	if err != nil {
		result.Status = CatalogStatusInvalid
		result.Reason = fmt.Sprintf("invalid source: %s", err)
		w.applyTransition(entry, result.Status, result.Reason)
		return result
	}

	// Step 2: Check repo accessibility (discover remote roles)
	roles, err := w.Client.DiscoverRemoteRoles(ctx, source)
	if err != nil {
		if isUnreachableError(err) {
			result.Status = CatalogStatusUnreachable
			result.Reason = fmt.Sprintf("repo unreachable: %s", err)
		} else {
			result.Status = CatalogStatusUnreachable
			result.Reason = fmt.Sprintf("discovery failed: %s", err)
		}
		result.Err = err
		w.applyTransition(entry, result.Status, result.Reason)
		return result
	}

	// Step 3: Find the specific role in the discovered roles
	var found *RoleRepoRemoteRole
	for idx := range roles {
		if roles[idx].Candidate.Name == entry.Name && roles[idx].Candidate.RolePath == entry.RolePath {
			found = &roles[idx]
			break
		}
	}
	if found == nil {
		result.Status = CatalogStatusInvalid
		result.Reason = fmt.Sprintf("role %q not found at path %q in %s", entry.Name, entry.RolePath, source.FullName())
		w.applyTransition(entry, result.Status, result.Reason)
		return result
	}

	// Step 4: Validate role.yaml structure by fetching and parsing it
	if err := w.validateRoleYAML(ctx, source, *found); err != nil {
		result.Status = CatalogStatusInvalid
		result.Reason = fmt.Sprintf("role.yaml validation failed: %s", err)
		result.Err = err
		w.applyTransition(entry, result.Status, result.Reason)
		return result
	}

	// Step 5: All checks passed → verified
	result.Status = CatalogStatusVerified
	result.Reason = "all checks passed"
	entry.FolderHash = found.FolderHash
	w.applyTransition(entry, result.Status, result.Reason)
	return result
}

func (w *NormalizeWorker) applyTransition(entry *RoleRepoCatalogEntry, status RoleRepoCatalogStatus, reason string) {
	now := w.NowFn().UTC()
	entry.Status = status
	entry.StatusReason = reason
	entry.UpdatedAt = now
	if status == CatalogStatusVerified {
		entry.VerifiedAt = &now
	}
}

// validateRoleYAML fetches and parses the role.yaml to ensure it has required fields.
func (w *NormalizeWorker) validateRoleYAML(ctx context.Context, source RoleRepoSource, remote RoleRepoRemoteRole) error {
	// Find the role.yaml blob SHA
	var yamlSHA string
	for _, f := range remote.Files {
		if f.Type == "blob" && strings.HasSuffix(f.Path, "role.yaml") {
			yamlSHA = f.SHA
			break
		}
	}
	if yamlSHA == "" {
		return fmt.Errorf("role.yaml not found in file tree")
	}

	data, err := w.Client.getBlobContent(ctx, source, yamlSHA)
	if err != nil {
		return fmt.Errorf("fetch role.yaml: %w", err)
	}

	var roleYAML struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal(data, &roleYAML); err != nil {
		return fmt.Errorf("parse role.yaml: %w", err)
	}
	if strings.TrimSpace(roleYAML.Name) == "" {
		return fmt.Errorf("role.yaml missing required field: name")
	}
	if strings.TrimSpace(roleYAML.Description) == "" {
		return fmt.Errorf("role.yaml missing required field: description")
	}
	return nil
}

// isUnreachableError checks if the error indicates a network/auth issue rather than data issue.
func isUnreachableError(err error) bool {
	if ghErr, ok := err.(*RoleRepoGitHubError); ok {
		return ghErr.StatusCode == 404 || ghErr.IsAuthOrRateLimit() || ghErr.StatusCode >= 500
	}
	return true // non-GitHub errors (DNS, timeout) are unreachable
}

// CatalogFromDiscoveredRoles creates catalog entries from discovered remote roles.
func CatalogFromDiscoveredRoles(roles []RoleRepoRemoteRole, nowFn func() time.Time) []RoleRepoCatalogEntry {
	now := nowFn().UTC()
	entries := make([]RoleRepoCatalogEntry, 0, len(roles))
	for _, role := range roles {
		entries = append(entries, RoleRepoCatalogEntry{
			Name:         role.Candidate.Name,
			Source:       role.Candidate.Source.FullName(),
			SourceType:   role.Candidate.SourceType,
			SourceURL:    role.Candidate.SourceURL,
			RolePath:     role.Candidate.RolePath,
			FolderHash:   role.FolderHash,
			Status:       CatalogStatusDiscovered,
			DiscoveredAt: now,
			UpdatedAt:    now,
		})
	}
	return entries
}

// FormatNormalizeResults produces a human-readable summary of normalization results.
func FormatNormalizeResults(results []NormalizeResult) string {
	if len(results) == 0 {
		return "No discovered roles to normalize.\n"
	}
	var b strings.Builder
	var verified, invalid, unreachable int
	for _, r := range results {
		switch r.Status {
		case CatalogStatusVerified:
			verified++
		case CatalogStatusInvalid:
			invalid++
		case CatalogStatusUnreachable:
			unreachable++
		}
		fmt.Fprintf(&b, "- %s/%s: %s (%s)\n", r.Entry.Source, r.Entry.Name, r.Status, r.Reason)
	}
	fmt.Fprintf(&b, "\nSummary: %d verified, %d invalid, %d unreachable\n", verified, invalid, unreachable)
	return b.String()
}
