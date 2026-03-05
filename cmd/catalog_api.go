package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newCatalogServeCmd() *cobra.Command {
	var addr string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve catalog HTTP APIs",
		Long:  "Serve catalog list/search/detail/repo/stats APIs for frontend consumption.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := GetApp(cmd)
			catalogPath := internal.ResolveCatalogPath(app.Git.Root())
			handler := internal.NewCatalogAPIHandler(catalogPath)
			server := &http.Server{
				Addr:              addr,
				Handler:           handler,
				ReadHeaderTimeout: 5 * time.Second,
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Catalog API listening on %s (base: %s)\n", addr, "/api")
			return server.ListenAndServe()
		},
	}
	cmd.Flags().StringVar(&addr, "addr", ":8787", "HTTP listen address")
	return cmd
}
