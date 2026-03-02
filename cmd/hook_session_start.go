// cmd/hook_session_start.go
package cmd

import (
	"fmt"
	"os"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newHookSessionStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "session-start",
		Short: "Handle session start event (role injection)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHookSessionStart(cmd)
		},
	}
}

func runHookSessionStart(cmd *cobra.Command) error {
	providerStr, _ := cmd.Flags().GetString("provider")

	input, err := internal.ParseHookInput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] session-start: parse input: %v\n", err)
		return nil // non-fatal
	}

	// Resolve provider
	if providerStr != "" && providerStr != "auto" {
		input.Provider = internal.ParseProvider(providerStr)
	} else {
		input.Provider = internal.DetectProvider()
	}

	// Detect worktree
	wt, err := internal.ResolveWorktree(input.CWD)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] session-start: resolve worktree: %v\n", err)
		return nil
	}
	if wt == nil {
		// Not in a managed worktree — bridge skills for providers that need it
		if input.Provider == internal.ProviderOpenCode {
			if err := internal.BridgeSkillsForProvider(input.CWD, string(input.Provider)); err != nil {
				fmt.Fprintf(os.Stderr, "[agent-team] session-start: bridge skills: %v\n", err)
			}
		}
		fmt.Fprintf(os.Stderr, "[agent-team] Ready. Use 'agent-team worker create <role>' to start a worker session.\n")
		return nil
	}

	fmt.Fprintf(os.Stderr, "[agent-team] session-start: worktree=%s worker=%s\n", wt.WtPath, wt.WorkerID)

	// Load worker config
	cfg, err := internal.LoadWorkerFromWorktree(wt.WtPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] session-start: load worker config: %v\n", err)
		return nil
	}

	// Inject role prompt
	if cfg.RolePath != "" {
		// Global role — use explicit path
		err = internal.InjectRolePromptWithPath(wt.WtPath, wt.WorkerID, cfg.Role, cfg.RolePath, wt.MainRoot)
	} else {
		// Project role — resolve from .agents/teams/
		err = internal.InjectRolePrompt(wt.WtPath, wt.WorkerID, cfg.Role, wt.MainRoot)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] session-start: inject role prompt: %v\n", err)
		return nil
	}

	fmt.Fprintf(os.Stderr, "[agent-team] session-start: role '%s' injected for worker '%s'\n", cfg.Role, wt.WorkerID)
	return nil
}
