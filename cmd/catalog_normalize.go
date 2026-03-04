package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newCatalogNormalizeCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "normalize",
		Short: "Validate discovered roles and update their status",
		Long:  "Run the normalize worker to transition discovered roles to verified, invalid, or unreachable based on validation checks.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunCatalogNormalize(jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func (a *App) RunCatalogNormalize(jsonOut bool) error {
	catalogPath := internal.ResolveCatalogPath(a.Git.Root())
	catalog, err := internal.ReadRoleRepoCatalog(catalogPath)
	if err != nil {
		return fmt.Errorf("read catalog: %w", err)
	}

	discovered := internal.FilterCatalogByStatus(catalog, internal.CatalogStatusDiscovered)
	if len(discovered) == 0 {
		if jsonOut {
			fmt.Println("[]")
		} else {
			fmt.Println("No discovered roles to normalize.")
		}
		return nil
	}

	client := internal.NewRoleRepoGitHubClient()
	worker := internal.NewNormalizeWorker(client)
	results := worker.NormalizeAll(context.Background(), &catalog)

	if err := internal.WriteRoleRepoCatalog(catalogPath, catalog); err != nil {
		return fmt.Errorf("write catalog: %w", err)
	}

	if jsonOut {
		type resultJSON struct {
			Name   string `json:"name"`
			Source string `json:"source"`
			Status string `json:"status"`
			Reason string `json:"reason"`
			Error  string `json:"error,omitempty"`
		}
		out := make([]resultJSON, 0, len(results))
		for _, r := range results {
			item := resultJSON{
				Name:   r.Entry.Name,
				Source: r.Entry.Source,
				Status: string(r.Status),
				Reason: r.Reason,
			}
			if r.Err != nil {
				item.Error = r.Err.Error()
			}
			out = append(out, item)
		}
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Print(internal.FormatNormalizeResults(results))
	return nil
}

func newCatalogDiscoverCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "discover <owner/repo>",
		Short: "Discover roles from a GitHub repository and add to catalog",
		Long:  "Scan a GitHub repository for role definitions and add any new ones to the catalog with 'discovered' status.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunCatalogDiscover(args[0], jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func (a *App) RunCatalogDiscover(sourceArg string, jsonOut bool) error {
	source, err := internal.ParseRoleRepoSource(sourceArg)
	if err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	client := internal.NewRoleRepoGitHubClient()
	roles, err := client.DiscoverRemoteRoles(context.Background(), source)
	if err != nil {
		return fmt.Errorf("discover roles: %w", err)
	}

	if len(roles) == 0 {
		if jsonOut {
			fmt.Println("[]")
		} else {
			fmt.Printf("No roles found in %s.\n", source.FullName())
		}
		return nil
	}

	catalogPath := internal.ResolveCatalogPath(a.Git.Root())
	catalog, err := internal.ReadRoleRepoCatalog(catalogPath)
	if err != nil {
		return fmt.Errorf("read catalog: %w", err)
	}

	newEntries := internal.CatalogFromDiscoveredRoles(roles, time.Now)
	var added int
	for _, entry := range newEntries {
		if _, exists := internal.FindCatalogEntry(catalog, entry.Source, entry.Name); exists {
			continue
		}
		internal.UpsertCatalogEntry(&catalog, entry)
		added++
	}

	if err := internal.WriteRoleRepoCatalog(catalogPath, catalog); err != nil {
		return fmt.Errorf("write catalog: %w", err)
	}

	if jsonOut {
		output := map[string]any{
			"source":     source.FullName(),
			"discovered": len(roles),
			"added":      added,
			"skipped":    len(roles) - added,
		}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Discovered %d role(s) in %s, added %d new to catalog (%d already known).\n",
		len(roles), source.FullName(), added, len(roles)-added)
	if added > 0 {
		fmt.Println("\nRun 'agent-team catalog normalize' to validate discovered roles.")
	}
	return nil
}
