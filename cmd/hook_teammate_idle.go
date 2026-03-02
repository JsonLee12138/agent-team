// cmd/hook_teammate_idle.go
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newHookTeammateIdleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "teammate-idle",
		Short: "Handle teammate idle event (notify main controller)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHookTeammateIdle(cmd)
		},
	}
}

func runHookTeammateIdle(cmd *cobra.Command) error {
	input, err := internal.ParseHookInput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] teammate-idle: parse input: %v\n", err)
		return nil
	}

	wt, err := internal.ResolveWorktree(input.CWD)
	if err != nil || wt == nil {
		return nil
	}

	msg := fmt.Sprintf("Worker idle: %s", wt.WorkerID)
	notifyCmd := exec.Command("agent-team", "reply-main", msg)
	notifyCmd.Dir = wt.WtPath
	notifyCmd.Stderr = os.Stderr
	if err := notifyCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] teammate-idle: notify main failed: %v\n", err)
	}

	return nil
}
