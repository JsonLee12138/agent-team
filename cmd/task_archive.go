package cmd

import (
	"fmt"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newTaskArchiveCmd() *cobra.Command {
	var mergedSHA string
	var strict bool
	cmd := &cobra.Command{
		Use:   "archive <task-id> --merged-sha <sha>",
		Short: "Archive a verifying task after merge",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunTaskArchive(args[0], mergedSHA, strict)
		},
	}
	cmd.Flags().StringVar(&mergedSHA, "merged-sha", "", "Merged commit SHA")
	cmd.Flags().BoolVar(&strict, "strict", false, "Require verification result 'pass' before archive")
	_ = cmd.MarkFlagRequired("merged-sha")
	return cmd
}

func (a *App) RunTaskArchive(taskID, mergedSHA string, strict bool) error {
	record, err := internal.ArchiveTask(a.Git.Root(), taskID, mergedSHA, strict, time.Now().UTC())
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
