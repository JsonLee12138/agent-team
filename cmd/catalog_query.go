package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newCatalogListCmd() *cobra.Command {
	var status string
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List roles in the catalog",
		Long:  "List cataloged roles. By default only 'verified' roles are shown. Use --status to filter by status (discovered, verified, invalid, unreachable, all).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunCatalogList(status, jsonOut)
		},
	}
	cmd.Flags().StringVar(&status, "status", "verified", "Filter by status (discovered, verified, invalid, unreachable, all)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func (a *App) RunCatalogList(status string, jsonOut bool) error {
	catalogPath := internal.ResolveCatalogPath(a.Git.Root())
	catalog, err := internal.ReadRoleRepoCatalog(catalogPath)
	if err != nil {
		return fmt.Errorf("read catalog: %w", err)
	}

	var filterStatus internal.RoleRepoCatalogStatus
	if status != "all" && status != "" {
		filterStatus = internal.RoleRepoCatalogStatus(status)
		if !internal.ValidCatalogStatuses[filterStatus] {
			return fmt.Errorf("invalid status %q; use: discovered, verified, invalid, unreachable, all", status)
		}
	}

	entries := internal.FilterCatalogByStatus(catalog, filterStatus)

	if jsonOut {
		if entries == nil {
			entries = []internal.RoleRepoCatalogEntry{}
		}
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	if len(entries) == 0 {
		if filterStatus != "" {
			fmt.Printf("No %s roles in the catalog.\n", filterStatus)
		} else {
			fmt.Println("Catalog is empty.")
		}
		return nil
	}

	fmt.Printf("%-20s %-24s %-12s %-12s %s\n", "ROLE", "SOURCE", "STATUS", "INSTALLS", "UPDATED")
	fmt.Printf("%-20s %-24s %-12s %-12s %s\n",
		"────────────────────", "────────────────────────", "────────────", "────────────", "────────────")
	for _, e := range entries {
		updated := "-"
		if !e.UpdatedAt.IsZero() {
			updated = e.UpdatedAt.Format("2006-01-02")
		}
		fmt.Printf("%-20s %-24s %-12s %-12d %s\n", e.Name, e.Source, e.Status, e.InstallCount, updated)
	}
	fmt.Printf("\nTotal: %d role(s)\n", len(entries))
	return nil
}

func newCatalogSearchCmd() *cobra.Command {
	var status string
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search roles in the catalog",
		Long:  "Search cataloged roles by name or source. By default only 'verified' roles are shown.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunCatalogSearch(args[0], status, jsonOut)
		},
	}
	cmd.Flags().StringVar(&status, "status", "verified", "Filter by status (discovered, verified, invalid, unreachable, all)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func (a *App) RunCatalogSearch(query, status string, jsonOut bool) error {
	catalogPath := internal.ResolveCatalogPath(a.Git.Root())
	catalog, err := internal.ReadRoleRepoCatalog(catalogPath)
	if err != nil {
		return fmt.Errorf("read catalog: %w", err)
	}

	var filterStatus internal.RoleRepoCatalogStatus
	if status != "all" && status != "" {
		filterStatus = internal.RoleRepoCatalogStatus(status)
		if !internal.ValidCatalogStatuses[filterStatus] {
			return fmt.Errorf("invalid status %q; use: discovered, verified, invalid, unreachable, all", status)
		}
	}

	entries := internal.SearchCatalog(catalog, query, filterStatus)

	if jsonOut {
		if entries == nil {
			entries = []internal.RoleRepoCatalogEntry{}
		}
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	if len(entries) == 0 {
		fmt.Printf("No roles matching %q found.\n", query)
		return nil
	}

	fmt.Printf("Search results for %q:\n\n", query)
	fmt.Printf("  %-20s %-24s %-12s %s\n", "ROLE", "SOURCE", "STATUS", "INSTALLS")
	fmt.Printf("  %-20s %-24s %-12s %s\n",
		"────────────────────", "────────────────────────", "────────────", "────────────")
	for _, e := range entries {
		fmt.Printf("  %-20s %-24s %-12s %d\n", e.Name, e.Source, e.Status, e.InstallCount)
	}
	fmt.Printf("\n%d result(s)\n", len(entries))
	return nil
}

func newCatalogShowCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "show <role-name>",
		Short: "Show detailed information about a cataloged role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunCatalogShow(args[0], jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func (a *App) RunCatalogShow(name string, jsonOut bool) error {
	catalogPath := internal.ResolveCatalogPath(a.Git.Root())
	catalog, err := internal.ReadRoleRepoCatalog(catalogPath)
	if err != nil {
		return fmt.Errorf("read catalog: %w", err)
	}

	entry, found := internal.FindCatalogEntryByName(catalog, name)
	if !found {
		return fmt.Errorf("role %q not found in catalog", name)
	}

	if jsonOut {
		data, err := json.MarshalIndent(entry, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Role: %s\n", entry.Name)
	fmt.Printf("Source: %s\n", entry.Source)
	fmt.Printf("Source URL: %s\n", entry.SourceURL)
	fmt.Printf("Role Path: %s\n", entry.RolePath)
	fmt.Printf("Status: %s\n", entry.Status)
	if entry.StatusReason != "" {
		fmt.Printf("Status Reason: %s\n", entry.StatusReason)
	}
	fmt.Printf("Folder Hash: %s\n", entry.FolderHash)
	fmt.Printf("Installs: %d\n", entry.InstallCount)
	fmt.Printf("Discovered: %s\n", entry.DiscoveredAt.Format("2006-01-02 15:04:05"))
	if entry.VerifiedAt != nil {
		fmt.Printf("Verified: %s\n", entry.VerifiedAt.Format("2006-01-02 15:04:05"))
	}
	if !entry.UpdatedAt.IsZero() {
		fmt.Printf("Updated: %s\n", entry.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

func newCatalogRepoCmd() *cobra.Command {
	var status string
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "repo <owner/repo>",
		Short: "Show all roles from a specific repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunCatalogRepo(args[0], status, jsonOut)
		},
	}
	cmd.Flags().StringVar(&status, "status", "verified", "Filter by status (discovered, verified, invalid, unreachable, all)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func (a *App) RunCatalogRepo(repo, status string, jsonOut bool) error {
	catalogPath := internal.ResolveCatalogPath(a.Git.Root())
	catalog, err := internal.ReadRoleRepoCatalog(catalogPath)
	if err != nil {
		return fmt.Errorf("read catalog: %w", err)
	}

	var filterStatus internal.RoleRepoCatalogStatus
	if status != "all" && status != "" {
		filterStatus = internal.RoleRepoCatalogStatus(status)
		if !internal.ValidCatalogStatuses[filterStatus] {
			return fmt.Errorf("invalid status %q", status)
		}
	}

	entries := internal.FilterCatalogByStatus(catalog, filterStatus)
	var repoEntries []internal.RoleRepoCatalogEntry
	for _, e := range entries {
		if strings.EqualFold(e.Source, repo) {
			repoEntries = append(repoEntries, e)
		}
	}

	if jsonOut {
		if repoEntries == nil {
			repoEntries = []internal.RoleRepoCatalogEntry{}
		}
		data, err := json.MarshalIndent(repoEntries, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	if len(repoEntries) == 0 {
		fmt.Printf("No roles from %s in the catalog.\n", repo)
		return nil
	}

	fmt.Printf("Roles from %s:\n\n", repo)
	fmt.Printf("  %-20s %-12s %-12s %s\n", "ROLE", "STATUS", "INSTALLS", "PATH")
	fmt.Printf("  %-20s %-12s %-12s %s\n",
		"────────────────────", "────────────", "────────────", "────────────────────────")
	for _, e := range repoEntries {
		fmt.Printf("  %-20s %-12s %-12d %s\n", e.Name, e.Status, e.InstallCount, e.RolePath)
	}
	fmt.Printf("\n%d role(s)\n", len(repoEntries))
	return nil
}

func newCatalogStatsCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show catalog statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunCatalogStats(jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func (a *App) RunCatalogStats(jsonOut bool) error {
	catalogPath := internal.ResolveCatalogPath(a.Git.Root())
	catalog, err := internal.ReadRoleRepoCatalog(catalogPath)
	if err != nil {
		return fmt.Errorf("read catalog: %w", err)
	}

	stats := internal.CatalogStats(catalog)
	repos := internal.GroupCatalogByRepo(catalog.Entries)

	if jsonOut {
		output := map[string]any{
			"total":        len(catalog.Entries),
			"byStatus":     stats,
			"repositories": len(repos),
		}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Catalog Statistics:\n\n")
	fmt.Printf("  Total roles:    %d\n", len(catalog.Entries))
	fmt.Printf("  Repositories:   %d\n", len(repos))
	fmt.Printf("  Verified:       %d\n", stats[internal.CatalogStatusVerified])
	fmt.Printf("  Discovered:     %d\n", stats[internal.CatalogStatusDiscovered])
	fmt.Printf("  Invalid:        %d\n", stats[internal.CatalogStatusInvalid])
	fmt.Printf("  Unreachable:    %d\n", stats[internal.CatalogStatusUnreachable])
	return nil
}
