package internal

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRoleRepoCatalogRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.json")
	now := time.Now().UTC().Round(time.Second)

	catalog := RoleRepoCatalog{
		Version: RoleRepoCatalogVersion,
		Entries: []RoleRepoCatalogEntry{
			{
				Name:         "frontend",
				Source:       "acme/roles",
				SourceType:   "github",
				SourceURL:    "https://github.com/acme/roles",
				RolePath:     "skills/frontend",
				FolderHash:   "abc123",
				Status:       CatalogStatusDiscovered,
				DiscoveredAt: now,
				UpdatedAt:    now,
			},
		},
	}

	if err := WriteRoleRepoCatalog(path, catalog); err != nil {
		t.Fatalf("WriteRoleRepoCatalog: %v", err)
	}

	got, err := ReadRoleRepoCatalog(path)
	if err != nil {
		t.Fatalf("ReadRoleRepoCatalog: %v", err)
	}
	if len(got.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got.Entries))
	}
	if got.Entries[0].Name != "frontend" {
		t.Fatalf("expected name=frontend, got %s", got.Entries[0].Name)
	}
	if got.Entries[0].Status != CatalogStatusDiscovered {
		t.Fatalf("expected status=discovered, got %s", got.Entries[0].Status)
	}
}

func TestRoleRepoCatalogMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.json")
	got, err := ReadRoleRepoCatalog(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Entries) != 0 {
		t.Fatalf("expected empty entries, got %d", len(got.Entries))
	}
	if got.Version != RoleRepoCatalogVersion {
		t.Fatalf("expected version %d, got %d", RoleRepoCatalogVersion, got.Version)
	}
}

func TestRoleRepoCatalogCorrupt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.json")
	if err := os.WriteFile(path, []byte("{bad-json"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := ReadRoleRepoCatalog(path)
	if err == nil {
		t.Fatal("expected error for corrupt file")
	}
	if len(got.Entries) != 0 {
		t.Fatalf("expected empty recovered catalog, got %d entries", len(got.Entries))
	}
}

func TestUpsertCatalogEntry(t *testing.T) {
	catalog := RoleRepoCatalog{Version: RoleRepoCatalogVersion}
	now := time.Now().UTC()

	entry1 := RoleRepoCatalogEntry{
		Name:   "frontend",
		Source: "acme/roles",
		Status: CatalogStatusDiscovered,
	}
	UpsertCatalogEntry(&catalog, entry1)
	if len(catalog.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(catalog.Entries))
	}

	// Update existing entry
	entry2 := RoleRepoCatalogEntry{
		Name:      "frontend",
		Source:    "acme/roles",
		Status:    CatalogStatusVerified,
		UpdatedAt: now,
	}
	UpsertCatalogEntry(&catalog, entry2)
	if len(catalog.Entries) != 1 {
		t.Fatalf("expected 1 entry after upsert, got %d", len(catalog.Entries))
	}
	if catalog.Entries[0].Status != CatalogStatusVerified {
		t.Fatalf("expected status=verified, got %s", catalog.Entries[0].Status)
	}

	// Add different entry
	entry3 := RoleRepoCatalogEntry{
		Name:   "backend",
		Source: "acme/roles",
		Status: CatalogStatusDiscovered,
	}
	UpsertCatalogEntry(&catalog, entry3)
	if len(catalog.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(catalog.Entries))
	}
}

func TestFindCatalogEntry(t *testing.T) {
	catalog := RoleRepoCatalog{
		Entries: []RoleRepoCatalogEntry{
			{Name: "frontend", Source: "acme/roles"},
			{Name: "backend", Source: "acme/roles"},
			{Name: "frontend", Source: "other/repo"},
		},
	}

	// Find by source+name
	entry, found := FindCatalogEntry(catalog, "acme/roles", "frontend")
	if !found || entry.Name != "frontend" || entry.Source != "acme/roles" {
		t.Fatalf("expected to find acme/roles:frontend")
	}

	// Find by name only
	entry, found = FindCatalogEntryByName(catalog, "backend")
	if !found || entry.Name != "backend" {
		t.Fatalf("expected to find backend")
	}

	// Not found
	_, found = FindCatalogEntry(catalog, "acme/roles", "nonexistent")
	if found {
		t.Fatal("expected not found")
	}
}

func TestFilterCatalogByStatus(t *testing.T) {
	catalog := RoleRepoCatalog{
		Entries: []RoleRepoCatalogEntry{
			{Name: "a", Status: CatalogStatusDiscovered},
			{Name: "b", Status: CatalogStatusVerified},
			{Name: "c", Status: CatalogStatusVerified},
			{Name: "d", Status: CatalogStatusInvalid},
		},
	}

	// Filter verified
	verified := FilterCatalogByStatus(catalog, CatalogStatusVerified)
	if len(verified) != 2 {
		t.Fatalf("expected 2 verified, got %d", len(verified))
	}

	// Filter discovered
	discovered := FilterCatalogByStatus(catalog, CatalogStatusDiscovered)
	if len(discovered) != 1 {
		t.Fatalf("expected 1 discovered, got %d", len(discovered))
	}

	// All (empty status)
	all := FilterCatalogByStatus(catalog, "")
	if len(all) != 4 {
		t.Fatalf("expected 4 total, got %d", len(all))
	}
}

func TestSearchCatalog(t *testing.T) {
	catalog := RoleRepoCatalog{
		Entries: []RoleRepoCatalogEntry{
			{Name: "frontend-dev", Source: "acme/roles", Status: CatalogStatusVerified},
			{Name: "backend-api", Source: "acme/roles", Status: CatalogStatusVerified},
			{Name: "frontend-ui", Source: "other/repo", Status: CatalogStatusDiscovered},
		},
	}

	// Search by name, verified only
	results := SearchCatalog(catalog, "frontend", CatalogStatusVerified)
	if len(results) != 1 || results[0].Name != "frontend-dev" {
		t.Fatalf("expected 1 result (frontend-dev), got %v", results)
	}

	// Search by name, all statuses
	results = SearchCatalog(catalog, "frontend", "")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Search by source
	results = SearchCatalog(catalog, "acme", "")
	if len(results) != 2 {
		t.Fatalf("expected 2 results for acme, got %d", len(results))
	}

	// Case insensitive
	results = SearchCatalog(catalog, "FRONTEND", "")
	if len(results) != 2 {
		t.Fatalf("expected 2 results for FRONTEND, got %d", len(results))
	}
}

func TestValidateTransition(t *testing.T) {
	tests := []struct {
		from    RoleRepoCatalogStatus
		to      RoleRepoCatalogStatus
		wantErr bool
	}{
		{CatalogStatusDiscovered, CatalogStatusVerified, false},
		{CatalogStatusDiscovered, CatalogStatusInvalid, false},
		{CatalogStatusDiscovered, CatalogStatusUnreachable, false},
		{CatalogStatusInvalid, CatalogStatusDiscovered, false},
		{CatalogStatusUnreachable, CatalogStatusDiscovered, false},
		{CatalogStatusVerified, CatalogStatusDiscovered, false},
		// Disallowed
		{CatalogStatusVerified, CatalogStatusInvalid, true},
		{CatalogStatusInvalid, CatalogStatusVerified, true},
		{CatalogStatusUnreachable, CatalogStatusVerified, true},
	}

	for _, tt := range tests {
		err := ValidateTransition(tt.from, tt.to)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateTransition(%s, %s) = %v, wantErr=%v", tt.from, tt.to, err, tt.wantErr)
		}
	}
}

func TestCatalogStats(t *testing.T) {
	catalog := RoleRepoCatalog{
		Entries: []RoleRepoCatalogEntry{
			{Status: CatalogStatusVerified},
			{Status: CatalogStatusVerified},
			{Status: CatalogStatusDiscovered},
			{Status: CatalogStatusInvalid},
		},
	}

	stats := CatalogStats(catalog)
	if stats[CatalogStatusVerified] != 2 {
		t.Fatalf("expected 2 verified, got %d", stats[CatalogStatusVerified])
	}
	if stats[CatalogStatusDiscovered] != 1 {
		t.Fatalf("expected 1 discovered, got %d", stats[CatalogStatusDiscovered])
	}
	if stats[CatalogStatusInvalid] != 1 {
		t.Fatalf("expected 1 invalid, got %d", stats[CatalogStatusInvalid])
	}
}

func TestGroupCatalogByRepo(t *testing.T) {
	entries := []RoleRepoCatalogEntry{
		{Name: "a", Source: "acme/roles"},
		{Name: "b", Source: "acme/roles"},
		{Name: "c", Source: "other/repo"},
	}

	groups := GroupCatalogByRepo(entries)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if len(groups["acme/roles"]) != 2 {
		t.Fatalf("expected 2 in acme/roles, got %d", len(groups["acme/roles"]))
	}
}

func TestResolveCatalogPath(t *testing.T) {
	path := ResolveCatalogPath("/tmp/project")
	expected := filepath.Join("/tmp/project", ".agents", "catalog.json")
	if path != expected {
		t.Fatalf("expected %s, got %s", expected, path)
	}
}

func TestContainsInsensitive(t *testing.T) {
	tests := []struct {
		s, substr string
		want      bool
	}{
		{"Frontend-Dev", "frontend", true},
		{"frontend-dev", "FRONTEND", true},
		{"backend", "frontend", false},
		{"", "test", false},
		{"test", "", true},
	}
	for _, tt := range tests {
		got := containsInsensitive(tt.s, tt.substr)
		if got != tt.want {
			t.Errorf("containsInsensitive(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}
