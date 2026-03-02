package internal

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRoleRepoLockRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roles-lock.json")
	now := time.Now().UTC().Round(time.Second)

	lock := RoleRepoLockFile{Version: RoleRepoLockVersion, Entries: []RoleRepoLockEntry{{
		Name: "frontend", Source: "acme/roles", SourceType: "github", SourceURL: "https://github.com/acme/roles", RolePath: "skills/frontend", FolderHash: "abc", InstalledAt: now, UpdatedAt: now,
	}}}

	if err := WriteRoleRepoLock(path, lock); err != nil {
		t.Fatalf("WriteRoleRepoLock: %v", err)
	}

	got, err := ReadRoleRepoLock(path)
	if err != nil {
		t.Fatalf("ReadRoleRepoLock: %v", err)
	}
	if len(got.Entries) != 1 || got.Entries[0].Name != "frontend" {
		t.Fatalf("unexpected entries: %+v", got.Entries)
	}
}

func TestRoleRepoLockCorruptRecovery(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roles-lock.json")
	if err := os.WriteFile(path, []byte("{bad-json"), 0644); err != nil {
		t.Fatal(err)
	}

	lock, err := ReadRoleRepoLock(path)
	if err == nil {
		t.Fatal("expected corrupt error")
	}
	if !errors.Is(err, ErrRoleRepoLockCorrupt) {
		t.Fatalf("expected ErrRoleRepoLockCorrupt, got %v", err)
	}
	if len(lock.Entries) != 0 {
		t.Fatalf("expected empty recovered lock, got %+v", lock.Entries)
	}
}

func TestRoleRepoLockUpsertRemove(t *testing.T) {
	lock := RoleRepoLockFile{Version: RoleRepoLockVersion}
	now := time.Now().UTC().Round(time.Second)
	entry := RoleRepoLockEntry{Name: "frontend", UpdatedAt: now}
	UpsertRoleRepoLockEntry(&lock, entry)
	UpsertRoleRepoLockEntry(&lock, RoleRepoLockEntry{Name: "frontend", FolderHash: "v2", UpdatedAt: now})
	if len(lock.Entries) != 1 || lock.Entries[0].FolderHash != "v2" {
		t.Fatalf("upsert failed: %+v", lock.Entries)
	}
	removed := RemoveRoleRepoLockEntries(&lock, []string{"frontend"})
	if removed != 1 || len(lock.Entries) != 0 {
		t.Fatalf("remove failed: removed=%d entries=%d", removed, len(lock.Entries))
	}
}
