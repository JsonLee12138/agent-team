package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

var ErrRoleRepoLockCorrupt = errors.New("role-repo lock file is corrupt")

func ReadRoleRepoLock(path string) (RoleRepoLockFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return RoleRepoLockFile{Version: RoleRepoLockVersion, Entries: []RoleRepoLockEntry{}}, nil
		}
		return RoleRepoLockFile{}, err
	}

	var lock RoleRepoLockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		return RoleRepoLockFile{Version: RoleRepoLockVersion, Entries: []RoleRepoLockEntry{}}, fmt.Errorf("%w: %v", ErrRoleRepoLockCorrupt, err)
	}
	if lock.Version == 0 {
		lock.Version = RoleRepoLockVersion
	}
	if lock.Entries == nil {
		lock.Entries = []RoleRepoLockEntry{}
	}
	return lock, nil
}

func WriteRoleRepoLock(path string, lock RoleRepoLockFile) error {
	if lock.Version == 0 {
		lock.Version = RoleRepoLockVersion
	}
	if lock.Entries == nil {
		lock.Entries = []RoleRepoLockEntry{}
	}
	sort.Slice(lock.Entries, func(i, j int) bool {
		return lock.Entries[i].Name < lock.Entries[j].Name
	})
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func UpsertRoleRepoLockEntry(lock *RoleRepoLockFile, entry RoleRepoLockEntry) {
	if lock.Entries == nil {
		lock.Entries = []RoleRepoLockEntry{}
	}
	for i := range lock.Entries {
		if lock.Entries[i].Name == entry.Name {
			lock.Entries[i] = entry
			return
		}
	}
	lock.Entries = append(lock.Entries, entry)
}

func RemoveRoleRepoLockEntries(lock *RoleRepoLockFile, names []string) int {
	if len(names) == 0 || len(lock.Entries) == 0 {
		return 0
	}
	rm := map[string]bool{}
	for _, name := range names {
		rm[name] = true
	}
	before := len(lock.Entries)
	out := lock.Entries[:0]
	for _, entry := range lock.Entries {
		if !rm[entry.Name] {
			out = append(out, entry)
		}
	}
	lock.Entries = out
	return before - len(out)
}

func FindRoleRepoLockEntry(lock RoleRepoLockFile, name string) (RoleRepoLockEntry, bool) {
	for _, entry := range lock.Entries {
		if entry.Name == name {
			return entry, true
		}
	}
	return RoleRepoLockEntry{}, false
}
