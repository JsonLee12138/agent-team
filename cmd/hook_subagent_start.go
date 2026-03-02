// cmd/hook_subagent_start.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newHookSubagentStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "subagent-start",
		Short: "Handle subagent start event (inject parent context)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHookSubagentStart(cmd)
		},
	}
}

func runHookSubagentStart(cmd *cobra.Command) error {
	input, err := internal.ParseHookInput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] subagent-start: parse input: %v\n", err)
		return nil
	}

	// Determine parent worktree from parent_cwd or cwd
	parentCWD := input.ParentCWD
	if parentCWD == "" {
		parentCWD = input.CWD
	}

	wt, err := internal.ResolveWorktree(parentCWD)
	if err != nil || wt == nil {
		return nil
	}

	// Read CLAUDE.md from parent worktree and extract AGENT_TEAM section
	claudePath := filepath.Join(wt.WtPath, "CLAUDE.md")
	data, err := os.ReadFile(claudePath)
	if err != nil {
		// Try AGENTS.md as fallback
		agentsPath := filepath.Join(wt.WtPath, "AGENTS.md")
		data, err = os.ReadFile(agentsPath)
		if err != nil {
			return nil // no context file found
		}
	}

	section := internal.ExtractAgentTeamSection(string(data))
	if section != "" {
		fmt.Fprintf(os.Stderr, "%s\n", section)
	}

	return nil
}
