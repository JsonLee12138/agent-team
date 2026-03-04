package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// RoleRepoCatalogStatus represents the verification status of a cataloged role.
type RoleRepoCatalogStatus string

const (
	CatalogStatusDiscovered  RoleRepoCatalogStatus = "discovered"
	CatalogStatusVerified    RoleRepoCatalogStatus = "verified"
	CatalogStatusInvalid     RoleRepoCatalogStatus = "invalid"
	CatalogStatusUnreachable RoleRepoCatalogStatus = "unreachable"
)

// ValidCatalogStatuses is the set of allowed status values.
var ValidCatalogStatuses = map[RoleRepoCatalogStatus]bool{
	CatalogStatusDiscovered:  true,
	CatalogStatusVerified:    true,
	CatalogStatusInvalid:     true,
	CatalogStatusUnreachable: true,
}

// AllowedTransitions defines which status transitions are legal.
var AllowedTransitions = map[RoleRepoCatalogStatus][]RoleRepoCatalogStatus{
	CatalogStatusDiscovered:  {CatalogStatusVerified, CatalogStatusInvalid, CatalogStatusUnreachable},
	CatalogStatusInvalid:     {CatalogStatusDiscovered},
	CatalogStatusUnreachable: {CatalogStatusDiscovered},
	CatalogStatusVerified:    {CatalogStatusDiscovered},
}

// RoleRepoCatalogEntry is a single role tracked in the catalog.
type RoleRepoCatalogEntry struct {
	Name         string                `json:"name"`
	Source       string                `json:"source"`
	SourceType   string                `json:"sourceType"`
	SourceURL    string                `json:"sourceUrl"`
	RolePath     string                `json:"rolePath"`
	FolderHash   string                `json:"folderHash"`
	Status       RoleRepoCatalogStatus `json:"status"`
	StatusReason string                `json:"statusReason,omitempty"`
	DiscoveredAt time.Time             `json:"discoveredAt"`
	VerifiedAt   *time.Time            `json:"verifiedAt,omitempty"`
	UpdatedAt    time.Time             `json:"updatedAt"`
	InstallCount int                   `json:"installCount"`
}

// RoleRepoCatalog is the persisted catalog of all known roles.
type RoleRepoCatalog struct {
	Version int                    `json:"version"`
	Entries []RoleRepoCatalogEntry `json:"entries"`
}

const RoleRepoCatalogVersion = 1

// ValidateTransition checks whether transitioning from current to next is allowed.
func ValidateTransition(current, next RoleRepoCatalogStatus) error {
	allowed, ok := AllowedTransitions[current]
	if !ok {
		return fmt.Errorf("unknown current status %q", current)
	}
	for _, s := range allowed {
		if s == next {
			return nil
		}
	}
	return fmt.Errorf("transition %s → %s not allowed", current, next)
}

// ReadRoleRepoCatalog reads the catalog from disk, returning an empty catalog if not found.
func ReadRoleRepoCatalog(path string) (RoleRepoCatalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return RoleRepoCatalog{Version: RoleRepoCatalogVersion, Entries: []RoleRepoCatalogEntry{}}, nil
		}
		return RoleRepoCatalog{}, err
	}
	var catalog RoleRepoCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return RoleRepoCatalog{Version: RoleRepoCatalogVersion, Entries: []RoleRepoCatalogEntry{}},
			fmt.Errorf("catalog file corrupt: %w", err)
	}
	if catalog.Version == 0 {
		catalog.Version = RoleRepoCatalogVersion
	}
	if catalog.Entries == nil {
		catalog.Entries = []RoleRepoCatalogEntry{}
	}
	return catalog, nil
}

// WriteRoleRepoCatalog writes the catalog to disk.
func WriteRoleRepoCatalog(path string, catalog RoleRepoCatalog) error {
	if catalog.Version == 0 {
		catalog.Version = RoleRepoCatalogVersion
	}
	if catalog.Entries == nil {
		catalog.Entries = []RoleRepoCatalogEntry{}
	}
	sort.Slice(catalog.Entries, func(i, j int) bool {
		if catalog.Entries[i].Source == catalog.Entries[j].Source {
			return catalog.Entries[i].Name < catalog.Entries[j].Name
		}
		return catalog.Entries[i].Source < catalog.Entries[j].Source
	})
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

// UpsertCatalogEntry adds or updates an entry in the catalog by source+name key.
func UpsertCatalogEntry(catalog *RoleRepoCatalog, entry RoleRepoCatalogEntry) {
	if catalog.Entries == nil {
		catalog.Entries = []RoleRepoCatalogEntry{}
	}
	for i := range catalog.Entries {
		if catalog.Entries[i].Source == entry.Source && catalog.Entries[i].Name == entry.Name {
			catalog.Entries[i] = entry
			return
		}
	}
	catalog.Entries = append(catalog.Entries, entry)
}

// FindCatalogEntry looks up an entry by source and name.
func FindCatalogEntry(catalog RoleRepoCatalog, source, name string) (RoleRepoCatalogEntry, bool) {
	for _, entry := range catalog.Entries {
		if entry.Source == source && entry.Name == name {
			return entry, true
		}
	}
	return RoleRepoCatalogEntry{}, false
}

// FindCatalogEntryByName looks up an entry by name only (returns first match).
func FindCatalogEntryByName(catalog RoleRepoCatalog, name string) (RoleRepoCatalogEntry, bool) {
	for _, entry := range catalog.Entries {
		if entry.Name == name {
			return entry, true
		}
	}
	return RoleRepoCatalogEntry{}, false
}

// FilterCatalogByStatus returns entries matching the given status.
// If status is empty, returns all entries.
func FilterCatalogByStatus(catalog RoleRepoCatalog, status RoleRepoCatalogStatus) []RoleRepoCatalogEntry {
	if status == "" {
		return catalog.Entries
	}
	var result []RoleRepoCatalogEntry
	for _, entry := range catalog.Entries {
		if entry.Status == status {
			result = append(result, entry)
		}
	}
	return result
}

// SearchCatalog filters entries by name substring match and optional status.
func SearchCatalog(catalog RoleRepoCatalog, query string, status RoleRepoCatalogStatus) []RoleRepoCatalogEntry {
	entries := FilterCatalogByStatus(catalog, status)
	if query == "" {
		return entries
	}
	var result []RoleRepoCatalogEntry
	for _, entry := range entries {
		if containsInsensitive(entry.Name, query) || containsInsensitive(entry.Source, query) {
			result = append(result, entry)
		}
	}
	return result
}

// GroupCatalogByRepo groups entries by source repository.
func GroupCatalogByRepo(entries []RoleRepoCatalogEntry) map[string][]RoleRepoCatalogEntry {
	groups := map[string][]RoleRepoCatalogEntry{}
	for _, entry := range entries {
		groups[entry.Source] = append(groups[entry.Source], entry)
	}
	return groups
}

// CatalogStats returns counts by status.
func CatalogStats(catalog RoleRepoCatalog) map[RoleRepoCatalogStatus]int {
	stats := map[RoleRepoCatalogStatus]int{}
	for _, entry := range catalog.Entries {
		stats[entry.Status]++
	}
	return stats
}

// ResolveCatalogPath returns the catalog file path for a given project root.
func ResolveCatalogPath(root string) string {
	return filepath.Join(root, ".agents", "catalog.json")
}

func containsInsensitive(s, substr string) bool {
	sl := len(s)
	subl := len(substr)
	if subl > sl {
		return false
	}
	for i := 0; i <= sl-subl; i++ {
		match := true
		for j := 0; j < subl; j++ {
			sc := s[i+j]
			tc := substr[j]
			if sc >= 'A' && sc <= 'Z' {
				sc += 32
			}
			if tc >= 'A' && tc <= 'Z' {
				tc += 32
			}
			if sc != tc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
