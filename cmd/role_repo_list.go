package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newRoleRepoListCmd() *cobra.Command {
	var global bool
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed repository-managed roles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunRoleRepoList(global, jsonOut)
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Show only global role installs")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

type roleListEntry struct {
	Name      string `json:"name"`
	Source    string `json:"source"`
	RolePath  string `json:"rolePath"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

type roleListOutput struct {
	Project []roleListEntry `json:"project"`
	Global  []roleListEntry `json:"global"`
}

func (a *App) RunRoleRepoList(globalOnly bool, jsonOut bool) error {
	root := a.Git.Root()

	var projectEntries []roleListEntry
	var globalEntries []roleListEntry

	// Collect project roles (unless --global filter)
	if !globalOnly {
		projectEntries = collectScopeEntries(root, internal.RoleRepoScopeProject)
	}

	// Always collect global roles
	globalEntries = collectScopeEntries(root, internal.RoleRepoScopeGlobal)

	// JSON output
	if jsonOut {
		output := roleListOutput{
			Project: projectEntries,
			Global:  globalEntries,
		}
		if output.Project == nil {
			output.Project = []roleListEntry{}
		}
		if output.Global == nil {
			output.Global = []roleListEntry{}
		}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	// Table output
	totalProject := len(projectEntries)
	totalGlobal := len(globalEntries)
	total := totalProject + totalGlobal

	if total == 0 {
		fmt.Println("No installed roles found.")
		return nil
	}

	if !globalOnly && len(projectEntries) > 0 {
		fmt.Println("Project roles (.agent-team/teams/):")
		printRoleTable(projectEntries)
		fmt.Println()
	}

	if len(globalEntries) > 0 {
		fmt.Println("Global roles (~/.agents/roles/):")
		printRoleTable(globalEntries)
		fmt.Println()
	}

	if !globalOnly {
		fmt.Printf("Total: %d role(s) (%d project, %d global)\n", total, totalProject, totalGlobal)
	} else {
		fmt.Printf("Total: %d global role(s)\n", totalGlobal)
	}
	return nil
}

func collectScopeEntries(root string, scope internal.RoleRepoScope) []roleListEntry {
	installRoot, err := internal.ResolveRoleRepoInstallRoot(root, scope)
	if err != nil {
		return nil
	}
	names, err := internal.ListInstalledRoleRepoNames(installRoot)
	if err != nil {
		return nil
	}

	lockPath, _ := internal.ResolveRoleRepoLockPath(root, scope)
	lock, _ := internal.ReadRoleRepoLock(lockPath)

	sort.Strings(names)
	entries := make([]roleListEntry, 0, len(names))
	for _, name := range names {
		entry := roleListEntry{Name: name, Source: "(untracked)", RolePath: "-"}
		if lockEntry, ok := internal.FindRoleRepoLockEntry(lock, name); ok {
			entry.Source = lockEntry.Source
			entry.RolePath = lockEntry.RolePath
			if !lockEntry.UpdatedAt.IsZero() {
				entry.UpdatedAt = lockEntry.UpdatedAt.Format("2006-01-02")
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

func printRoleTable(entries []roleListEntry) {
	fmt.Printf("  %-20s %-28s %-12s %s\n", "ROLE", "SOURCE", "UPDATED", "PATH")
	fmt.Printf("  %-20s %-28s %-12s %s\n", "────────────────────", "────────────────────────────", "────────────", "────────────────────────")
	for _, e := range entries {
		updated := e.UpdatedAt
		if updated == "" {
			updated = "-"
		}
		fmt.Printf("  %-20s %-28s %-12s %s\n", e.Name, e.Source, updated, e.RolePath)
	}
}
