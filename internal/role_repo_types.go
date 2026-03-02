package internal

import "time"

// RoleRepoScope controls install scope for role-repo commands.
type RoleRepoScope string

const (
	RoleRepoScopeProject RoleRepoScope = "project"
	RoleRepoScopeGlobal  RoleRepoScope = "global"
)

// RoleRepoSource is a normalized remote source reference.
type RoleRepoSource struct {
	Original string
	Owner    string
	Repo     string
}

func (s RoleRepoSource) FullName() string {
	return s.Owner + "/" + s.Repo
}

func (s RoleRepoSource) HTTPSURL() string {
	return "https://github.com/" + s.FullName()
}

// RoleRepoCandidate identifies a role path in a source repository.
type RoleRepoCandidate struct {
	Name       string
	RolePath   string
	YAMLPath   string
	Source     RoleRepoSource
	SourceType string
	SourceURL  string
}

// RoleRepoSearchResult is one filtered GitHub search hit.
type RoleRepoSearchResult struct {
	Name      string
	Repo      string
	RolePath  string
	YAMLPath  string
	HTMLURL   string
	SourceURL string
}

// RoleRepoLockEntry tracks installation metadata for one role.
type RoleRepoLockEntry struct {
	Name        string    `json:"name"`
	Source      string    `json:"source"`
	SourceType  string    `json:"sourceType"`
	SourceURL   string    `json:"sourceUrl"`
	RolePath    string    `json:"rolePath"`
	FolderHash  string    `json:"folderHash"`
	InstalledAt time.Time `json:"installedAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// RoleRepoLockFile is persisted to project/global lock paths.
type RoleRepoLockFile struct {
	Version int                 `json:"version"`
	Entries []RoleRepoLockEntry `json:"entries"`
}

const RoleRepoLockVersion = 1

// RoleRepoCheckStatus summarizes check result for one installed role.
type RoleRepoCheckStatus struct {
	Name         string
	CurrentHash  string
	RemoteHash   string
	Source       string
	RolePath     string
	State        string // up_to_date | update_available | error
	Err          error
	RemoteExists bool
}
