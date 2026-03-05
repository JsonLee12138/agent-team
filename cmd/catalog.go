package cmd

import (
	"github.com/spf13/cobra"
)

func newCatalogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "Query and manage the role catalog",
		Long:  "Browse, search, and inspect verified roles in the catalog. By default only verified roles are shown.",
	}
	cmd.AddCommand(newCatalogListCmd())
	cmd.AddCommand(newCatalogSearchCmd())
	cmd.AddCommand(newCatalogShowCmd())
	cmd.AddCommand(newCatalogRepoCmd())
	cmd.AddCommand(newCatalogNormalizeCmd())
	cmd.AddCommand(newCatalogStatsCmd())
	cmd.AddCommand(newCatalogDiscoverCmd())
	cmd.AddCommand(newCatalogServeCmd())
	return cmd
}
