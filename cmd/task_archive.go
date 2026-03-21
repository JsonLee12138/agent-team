package cmd

import (
	"fmt"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newTaskArchiveCmd() *cobra.Command {
	var mergedSHA string
	cmd := &cobra.Command{
		Use:   "archive <task-id> --merged-sha <sha>",
		Short: "Archive a done task after merge",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunTaskArchive(args[0], mergedSHA)
		},
	}
	cmd.Flags().StringVar(&mergedSHA, "merged-sha", "", "Merged commit SHA")
	_ = cmd.MarkFlagRequired("merged-sha")
	return cmd
}

func (a *App) RunTaskArchive(taskID, mergedSHA string) error {
	record, err := internal.ArchiveTask(a.Git.Root(), taskID, mergedSHA, time.Now().UTC())
	if err != nil {
		return err
	}
	if record.WorkerID != "" {
		if err := a.RunWorkerDelete(record.WorkerID); err != nil {
			return fmt.Errorf("archive task '%s' but cleanup worker '%s' failed: %w", taskID, record.WorkerID, err)
		}
	}
	fmt.Printf("✓ Archived task '%s'\n", record.TaskID)
	return nil
}
