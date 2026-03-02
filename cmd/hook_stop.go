// cmd/hook_stop.go
package cmd

import (
	"fmt"
	"os"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newHookStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Handle stop/session-end event (warn about unarchived changes)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHookStop(cmd)
		},
	}
}

func runHookStop(cmd *cobra.Command) error {
	input, err := internal.ParseHookInput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] stop: parse input: %v\n", err)
		return nil
	}

	wt, err := internal.ResolveWorktree(input.CWD)
	if err != nil || wt == nil {
		return nil
	}

	active, err := internal.ListActiveChanges(wt.WtPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] stop: list changes: %v\n", err)
		return nil
	}

	if len(active) > 0 {
		fmt.Fprintf(os.Stderr, "[agent-team] stop: WARNING — %d unarchived change(s) in worker '%s':\n", len(active), wt.WorkerID)
		for _, c := range active {
			fmt.Fprintf(os.Stderr, "  - %s (status: %s)\n", c.Name, c.Status)
		}
		fmt.Fprintf(os.Stderr, "[agent-team] stop: Consider archiving before ending session.\n")
	}

	return nil
}
