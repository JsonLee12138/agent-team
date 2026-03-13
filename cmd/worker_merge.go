// cmd/worker_merge.go
package cmd

import (
	"fmt"
	"os"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newWorkerMergeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "merge <worker-id>",
		Short: "Merge a worker's branch into the current branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunWorkerMerge(args[0])
		},
	}
}

func (a *App) RunWorkerMerge(workerID string) error {
	root := a.Git.Root()
	wtPath := internal.WtPath(root, a.WtBase, workerID)
	branch := "team/" + workerID

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("worker '%s' not found", workerID)
	}

	mainBranch, _ := a.Git.CurrentBranch()

	fmt.Printf("Merging branch '%s' into '%s'...\n", branch, mainBranch)
	msg := fmt.Sprintf("merge: integrate work from worker '%s'", workerID)
	if err := a.Git.Merge(branch, msg); err != nil {
		return err
	}

	fmt.Printf("✓ Merged '%s' into %s\n", workerID, mainBranch)
	fmt.Printf("  → Run 'agent-team worker delete %s' to remove the worktree when done\n", workerID)
	return nil
}
