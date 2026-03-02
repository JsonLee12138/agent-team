// cmd/hook_post_tool_use.go
package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newHookPostToolUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "post-tool-use",
		Short: "Handle post-tool-use event (quality checks)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHookPostToolUse(cmd)
		},
	}
}

func runHookPostToolUse(cmd *cobra.Command) error {
	input, err := internal.ParseHookInput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] post-tool-use: parse input: %v\n", err)
		return nil
	}

	wt, err := internal.ResolveWorktree(input.CWD)
	if err != nil || wt == nil {
		return nil
	}

	// Load worker config to find role
	cfg, err := internal.LoadWorkerFromWorktree(wt.WtPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] post-tool-use: load worker config: %v\n", err)
		return nil
	}

	// Resolve role path
	var rolePath string
	if cfg.RolePath != "" {
		rolePath = cfg.RolePath
	} else {
		rolePath = internal.RoleDir(wt.MainRoot, cfg.Role)
	}

	// Read quality checks from role.yaml
	checks, err := internal.ReadRoleQualityChecks(rolePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] post-tool-use: read quality checks: %v\n", err)
		return nil
	}

	if len(checks) == 0 {
		return nil // no quality checks configured
	}

	// Execute each quality check with 15s timeout
	for _, check := range checks {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		bashCmd := exec.CommandContext(ctx, "bash", "-c", check)
		bashCmd.Dir = wt.WtPath
		var stdout, stderr bytes.Buffer
		bashCmd.Stdout = &stdout
		bashCmd.Stderr = &stderr

		err := bashCmd.Run()
		cancel()

		if err != nil {
			fmt.Fprintf(os.Stderr, "[agent-team] post-tool-use: FAIL — %s\n", check)
			if stderr.Len() > 0 {
				fmt.Fprintf(os.Stderr, "  stderr: %s\n", stderr.String())
			}
			if stdout.Len() > 0 {
				fmt.Fprintf(os.Stderr, "  stdout: %s\n", stdout.String())
			}
		} else {
			fmt.Fprintf(os.Stderr, "[agent-team] post-tool-use: PASS — %s\n", check)
		}
	}

	return nil
}
