// cmd/task_archive.go
package cmd

import (
	"fmt"
	"os"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newTaskArchiveCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:   "archive <worker-id> <change-name>",
		Short: "Archive a task change",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunTaskArchive(args[0], args[1], dir)
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "", "Override worker directory (for script use)")

	return cmd
}

func (a *App) RunTaskArchive(workerID, changeName, overrideDir string) error {
	root := a.Git.Root()

	// Allow script to override the worker directory
	var wtPath string
	if overrideDir != "" {
		wtPath = overrideDir
	} else {
		wtPath = internal.WtPath(root, a.WtBase, workerID)
	}

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("worker directory '%s' not found", wtPath)
	}

	change, err := internal.LoadChange(wtPath, changeName)
	if err != nil {
		return fmt.Errorf("load change: %w", err)
	}

	// Validate transition
	if err := internal.ValidateChangeTransition(change.Status, internal.ChangeStatusArchived); err != nil {
		return fmt.Errorf("cannot archive change in '%s' state: %w", change.Status, err)
	}

	// Apply transition
	if err := internal.ApplyChangeTransition(change, internal.ChangeStatusArchived); err != nil {
		return fmt.Errorf("apply transition: %w", err)
	}

	// Save change
	if err := internal.SaveChange(wtPath, change); err != nil {
		return fmt.Errorf("save change: %w", err)
	}

	fmt.Printf("✓ Archived change: %s\n", changeName)
	return nil
}
