// cmd/task_verify.go
package cmd

import (
	"fmt"
	"os"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newTaskVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify <worker-id> <change-name>",
		Short: "Run verification for a task change",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunTaskVerify(args[0], args[1])
		},
	}

	return cmd
}

func (a *App) RunTaskVerify(workerID, changeName string) error {
	root := a.Git.Root()
	wtPath := internal.WtPath(root, a.WtBase, workerID)

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("worker '%s' not found", workerID)
	}

	change, err := internal.LoadChange(wtPath, changeName)
	if err != nil {
		return fmt.Errorf("load change: %w", err)
	}

	// Check if change is in a state that can be verified
	if change.Status != internal.ChangeStatusImplementing && change.Status != internal.ChangeStatusVerifying {
		return fmt.Errorf("change must be in 'implementing' or 'verifying' state to verify, current: %s", change.Status)
	}

	// Transition to verifying if not already
	if change.Status == internal.ChangeStatusImplementing {
		if err := internal.ApplyChangeTransition(change, internal.ChangeStatusVerifying); err != nil {
			return fmt.Errorf("transition to verifying: %w", err)
		}
		if err := internal.SaveChange(wtPath, change); err != nil {
			return fmt.Errorf("save change: %w", err)
		}
		fmt.Printf("✓ Transitioned to verifying state\n")
	}

	// Run verification
	fmt.Printf("Running verification for '%s'...\n", changeName)
	result, err := internal.RunVerify(wtPath, change)
	if err != nil {
		return fmt.Errorf("verify: %w", err)
	}

	// Display result
	status := "✗ FAILED"
	if result.Passed {
		status = "✓ PASSED"
	}
	fmt.Printf("%s (exit code: %d, duration: %v)\n", status, result.ExitCode, result.Duration)
	if result.Output != "" {
		fmt.Printf("\nOutput:\n%s\n", result.Output)
	}

	// Auto-transition based on result
	if err := internal.AutoTransitionOnVerify(change, result.Passed); err != nil {
		return fmt.Errorf("auto-transition: %w", err)
	}

	if err := internal.SaveChange(wtPath, change); err != nil {
		return fmt.Errorf("save change: %w", err)
	}

	fmt.Printf("✓ Status updated to: %s\n", change.Status)
	return nil
}
