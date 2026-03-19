package cmd

import (
	"fmt"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newRecordMainPaneCmd() *cobra.Command {
	var root string

	cmd := &cobra.Command{
		Use:    "_record-main-pane",
		Hidden: true,
		Short:  "Record current main pane for this project (internal)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if root == "" {
				return fmt.Errorf("--root is required")
			}
			paneID, backend := detectCurrentPaneFromEnv()
			if paneID == "" {
				return nil
			}
			cfg := &internal.MainSessionConfig{
				PaneID:    paneID,
				Backend:   backend,
				UpdatedAt: time.Now().UTC().Format(time.RFC3339),
			}
			return cfg.Save(internal.MainSessionYAMLPath(root))
		},
	}

	cmd.Flags().StringVar(&root, "root", "", "Project root directory")
	return cmd
}
