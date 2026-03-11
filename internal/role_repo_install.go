package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var ErrRoleRepoInstallConflict = errors.New("role already exists")

func SelectRoleRepoRemotes(all []RoleRepoRemoteRole, roleNames []string) ([]RoleRepoRemoteRole, error) {
	if len(roleNames) == 0 {
		return all, nil
	}
	selected := make([]RoleRepoRemoteRole, 0, len(roleNames))
	index := map[string]RoleRepoRemoteRole{}
	for _, item := range all {
		index[item.Candidate.Name] = item
	}
	for _, roleName := range roleNames {
		item, ok := index[roleName]
		if !ok {
			return nil, fmt.Errorf("role %q not found in source", roleName)
		}
		selected = append(selected, item)
	}
	return selected, nil
}

func InstallRoleRepoRemoteRole(ctx context.Context, client *RoleRepoGitHubClient, remote RoleRepoRemoteRole, installRoot string, overwrite bool) (string, error) {
	targetDir := filepath.Join(installRoot, remote.Candidate.Name)
	if st, err := os.Stat(targetDir); err == nil && st.IsDir() {
		if !overwrite {
			return "", fmt.Errorf("%w: %s", ErrRoleRepoInstallConflict, targetDir)
		}
		if err := os.RemoveAll(targetDir); err != nil {
			return "", err
		}
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return "", err
		}
		if err := writeRoleRepoFiles(ctx, client, remote, targetDir); err != nil {
			return "", err
		}
		return targetDir, nil
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}
	if err := writeRoleRepoFiles(ctx, client, remote, targetDir); err != nil {
		return "", err
	}
	return targetDir, nil
}

func writeRoleRepoFiles(ctx context.Context, client *RoleRepoGitHubClient, remote RoleRepoRemoteRole, targetDir string) error {
	prefix := remote.Candidate.RolePath + "/"
	files := append([]roleRepoTreeEntry(nil), remote.Files...)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
	for _, file := range files {
		if file.Type != "blob" {
			continue
		}
		if !strings.HasPrefix(file.Path, prefix) {
			continue
		}
		rel := strings.TrimPrefix(file.Path, prefix)
		if rel == "" || strings.Contains(rel, "..") {
			return fmt.Errorf("invalid role file path: %s", file.Path)
		}
		data, err := client.getBlobContent(ctx, remote.Candidate.Source, file.SHA)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func ListInstalledRoleRepoNames(installRoot string) ([]string, error) {
	entries, err := os.ReadDir(installRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

func RemoveInstalledRoleRepo(installRoot string, names []string) (removed []string, missing []string, failed map[string]error) {
	failed = map[string]error{}
	for _, name := range names {
		target := filepath.Join(installRoot, name)
		st, err := os.Stat(target)
		if err != nil {
			if os.IsNotExist(err) {
				missing = append(missing, name)
				continue
			}
			failed[name] = err
			continue
		}
		if !st.IsDir() {
			failed[name] = fmt.Errorf("not a role directory: %s", target)
			continue
		}
		if err := os.RemoveAll(target); err != nil {
			failed[name] = err
			continue
		}
		removed = append(removed, name)
	}
	sort.Strings(removed)
	sort.Strings(missing)
	return removed, missing, failed
}
