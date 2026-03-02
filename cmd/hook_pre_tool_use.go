// cmd/hook_pre_tool_use.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newHookPreToolUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pre-tool-use",
		Short: "Handle pre-tool-use event (brainstorming gate)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHookPreToolUse(cmd)
		},
	}
}

func runHookPreToolUse(cmd *cobra.Command) error {
	input, err := internal.ParseHookInput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] pre-tool-use: parse input: %v\n", err)
		return nil
	}

	wt, err := internal.ResolveWorktree(input.CWD)
	if err != nil || wt == nil {
		return nil
	}

	// Check if .tasks/ directory exists
	tasksDir := internal.TasksDir(wt.WtPath)
	if _, err := os.Stat(tasksDir); os.IsNotExist(err) {
		return nil // no task system initialized, skip check
	}

	// Find active changes
	active, err := internal.ListActiveChanges(wt.WtPath)
	if err != nil || len(active) == 0 {
		return nil
	}

	// Check if the first active change has a design.md
	change := active[0]
	designPath := filepath.Join(internal.ChangeDirPath(wt.WtPath, change.Name), "design.md")
	if _, err := os.Stat(designPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "[agent-team] pre-tool-use: WARNING — No design.md found for change '%s'.\n", change.Name)
		fmt.Fprintf(os.Stderr, "[agent-team] pre-tool-use: Create a design document before writing code.\n")
		fmt.Fprintf(os.Stderr, "[agent-team] pre-tool-use: Path: %s\n", designPath)
	}

	return nil
}
