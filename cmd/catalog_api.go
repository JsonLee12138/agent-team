package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newCatalogServeCmd() *cobra.Command {
	var addr string
	var normalizeInterval string
	var visibilityRefreshInterval string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve catalog HTTP APIs",
		Long:  "Serve catalog list/search/detail/repo/stats APIs for frontend consumption.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := GetApp(cmd)
			catalogPath := internal.ResolveCatalogPath(app.Git.Root())
			cfg, err := internal.EnsureCatalogPipelineConfig(app.Git.Root())
			if err != nil {
				return err
			}
			if normalizeInterval != "" {
				cfg.Normalize.Interval = normalizeInterval
			}
			if visibilityRefreshInterval != "" {
				cfg.Visibility.RefreshInterval = visibilityRefreshInterval
			}

			normalizeEvery, err := cfg.NormalizeIntervalDuration()
			if err != nil {
				return err
			}
			visibilityEvery, err := cfg.VisibilityRefreshIntervalDuration()
			if err != nil {
				return err
			}
			if normalizeEvery < 0 || visibilityEvery < 0 {
				return fmt.Errorf("intervals must be >= 0")
			}

			store := internal.NewCatalogStore(catalogPath, visibilityEvery)
			handler := internal.NewCatalogAPIHandlerWithStore(store)

			if normalizeEvery > 0 {
				go runCatalogNormalizeLoop(cmd.Context(), cmd.OutOrStdout(), catalogPath, store, normalizeEvery)
			}

			server := &http.Server{
				Addr:              addr,
				Handler:           handler,
				ReadHeaderTimeout: 5 * time.Second,
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Catalog API listening on %s (base: %s)\n", addr, "/api")
			fmt.Fprintf(cmd.OutOrStdout(), "Normalize interval: %s | Visibility refresh: %s\n", normalizeEvery, visibilityEvery)
			return server.ListenAndServe()
		},
	}
	cmd.Flags().StringVar(&addr, "addr", ":8787", "HTTP listen address")
	cmd.Flags().StringVar(&normalizeInterval, "normalize-interval", "", "Polling interval for normalization (0 to disable)")
	cmd.Flags().StringVar(&visibilityRefreshInterval, "visibility-refresh-interval", "", "Catalog refresh interval for API cache (0 to disable caching)")
	return cmd
}

func runCatalogNormalizeLoop(ctx context.Context, out io.Writer, catalogPath string, store *internal.CatalogStore, interval time.Duration) {
	normalizeOnce(ctx, out, catalogPath, store)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			normalizeOnce(ctx, out, catalogPath, store)
		}
	}
}

func normalizeOnce(ctx context.Context, out io.Writer, catalogPath string, store *internal.CatalogStore) {
	catalog, err := internal.ReadRoleRepoCatalog(catalogPath)
	if err != nil {
		fmt.Fprintf(out, "normalize: read catalog failed: %v\n", err)
		return
	}
	worker := internal.NewNormalizeWorker(internal.NewRoleRepoGitHubClient())
	results := worker.NormalizeAll(ctx, &catalog)
	if len(results) == 0 {
		return
	}
	if err := internal.WriteRoleRepoCatalog(catalogPath, catalog); err != nil {
		fmt.Fprintf(out, "normalize: write catalog failed: %v\n", err)
		return
	}
	if store != nil {
		store.Set(catalog)
	}
	fmt.Fprint(out, internal.FormatNormalizeResults(results))
}
