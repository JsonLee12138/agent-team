package internal

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

func CheckRoleRepoUpdates(ctx context.Context, client *RoleRepoGitHubClient, installRoot string, lock RoleRepoLockFile) ([]RoleRepoCheckStatus, []string) {
	status := make([]RoleRepoCheckStatus, 0, len(lock.Entries))
	remoteCache := map[string]map[string]RoleRepoRemoteRole{}

	for _, entry := range lock.Entries {
		item := RoleRepoCheckStatus{
			Name:        entry.Name,
			CurrentHash: entry.FolderHash,
			Source:      entry.Source,
			RolePath:    entry.RolePath,
			State:       "error",
		}

		repoMap, ok := remoteCache[entry.Source]
		if !ok {
			source, err := ParseRoleRepoSource(entry.Source)
			if err != nil {
				item.Err = err
				status = append(status, item)
				continue
			}
			roles, err := client.DiscoverRemoteRoles(ctx, source)
			if err != nil {
				item.Err = err
				status = append(status, item)
				continue
			}
			repoMap = map[string]RoleRepoRemoteRole{}
			for _, role := range roles {
				repoMap[role.Candidate.RolePath] = role
			}
			remoteCache[entry.Source] = repoMap
		}

		remote, exists := repoMap[entry.RolePath]
		if !exists {
			item.RemoteExists = false
			item.Err = fmt.Errorf("role path %s not found in source", entry.RolePath)
			status = append(status, item)
			continue
		}

		item.RemoteExists = true
		item.RemoteHash = remote.FolderHash
		if remote.FolderHash != entry.FolderHash {
			item.State = "update_available"
		} else {
			item.State = "up_to_date"
		}
		status = append(status, item)
	}

	sort.Slice(status, func(i, j int) bool {
		return status[i].Name < status[j].Name
	})
	untracked := roleRepoUntrackedLocalRoles(installRoot, lock)
	return status, untracked
}

func roleRepoUntrackedLocalRoles(installRoot string, lock RoleRepoLockFile) []string {
	installed, err := ListInstalledRoleRepoNames(installRoot)
	if err != nil {
		return nil
	}
	tracked := map[string]bool{}
	for _, entry := range lock.Entries {
		tracked[entry.Name] = true
	}
	var out []string
	for _, name := range installed {
		if !tracked[name] {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

func UpdateRoleRepoFromLock(ctx context.Context, client *RoleRepoGitHubClient, installRoot string, lock *RoleRepoLockFile, overwrite bool, selectedNames []string, nowFn func() time.Time) (updated []string, skipped []string, failed map[string]error) {
	failed = map[string]error{}
	remoteCache := map[string]map[string]RoleRepoRemoteRole{}
	selectedSet := map[string]bool{}
	if len(selectedNames) > 0 {
		for _, name := range selectedNames {
			selectedSet[name] = true
		}
	}

	for i := range lock.Entries {
		entry := lock.Entries[i]
		if len(selectedSet) > 0 && !selectedSet[entry.Name] {
			skipped = append(skipped, entry.Name)
			continue
		}
		sourceRoles, ok := remoteCache[entry.Source]
		if !ok {
			source, err := ParseRoleRepoSource(entry.Source)
			if err != nil {
				failed[entry.Name] = err
				continue
			}
			roles, err := client.DiscoverRemoteRoles(ctx, source)
			if err != nil {
				failed[entry.Name] = err
				continue
			}
			sourceRoles = map[string]RoleRepoRemoteRole{}
			for _, role := range roles {
				sourceRoles[role.Candidate.RolePath] = role
			}
			remoteCache[entry.Source] = sourceRoles
		}

		remoteRole, ok := sourceRoles[entry.RolePath]
		if !ok {
			failed[entry.Name] = fmt.Errorf("role path %s not found in source", entry.RolePath)
			continue
		}
		if remoteRole.FolderHash == entry.FolderHash {
			skipped = append(skipped, entry.Name)
			continue
		}
		if !overwrite {
			skipped = append(skipped, entry.Name)
			continue
		}

		_, _, err := InstallRoleRepoRemoteRole(ctx, client, remoteRole, installRoot, true, nowFn)
		if err != nil {
			failed[entry.Name] = err
			continue
		}
		lock.Entries[i].Source = remoteRole.Candidate.Source.FullName()
		lock.Entries[i].SourceType = remoteRole.Candidate.SourceType
		lock.Entries[i].SourceURL = remoteRole.Candidate.SourceURL
		lock.Entries[i].RolePath = remoteRole.Candidate.RolePath
		lock.Entries[i].FolderHash = remoteRole.FolderHash
		lock.Entries[i].UpdatedAt = nowFn().UTC()
		updated = append(updated, entry.Name)
	}

	sort.Strings(updated)
	sort.Strings(skipped)
	return updated, skipped, failed
}

func FormatRoleRepoCheckSummary(statuses []RoleRepoCheckStatus, untracked []string) string {
	var b strings.Builder
	if len(statuses) == 0 {
		b.WriteString("No lock entries found.\n")
	} else {
		for _, st := range statuses {
			switch st.State {
			case "up_to_date":
				b.WriteString("- ")
				b.WriteString(st.Name)
				b.WriteString(": up to date\n")
			case "update_available":
				b.WriteString("- ")
				b.WriteString(st.Name)
				b.WriteString(": update available\n")
			default:
				b.WriteString("- ")
				b.WriteString(st.Name)
				b.WriteString(": error")
				if st.Err != nil {
					b.WriteString(" (")
					b.WriteString(st.Err.Error())
					b.WriteString(")")
				}
				b.WriteString("\n")
			}
		}
	}
	if len(untracked) > 0 {
		b.WriteString("Untracked local roles: ")
		b.WriteString(strings.Join(untracked, ", "))
		b.WriteString("\n")
	}
	return b.String()
}

func EnsureRoleRepoInstallRoot(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	return nil
}
