package cmd

import (
	"fmt"
	"os"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize agent-team for this project",
		Long: `Creates project-level structure: .agents/teams/, .agents/rules/ (with default rule files),
and root provider files (CLAUDE.md, AGENTS.md, GEMINI.md).

Idempotent: existing files are not overwritten; provider files only update the tagged section.

For global environment setup (provider detection, plugin role installation),
use 'agent-team setup' instead.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit()
		},
	}
	return cmd
}

func runInit() error {
	fmt.Println("agent-team init")
	fmt.Println()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Step 1: Initialize project structure (.agents/teams/)
	if err := internal.InitProject(cwd); err != nil {
		return err
	}
	fmt.Println("✓ .agents/teams/ ready")

	// Step 2: Initialize rules directory (.agents/rules/)
	created, err := internal.InitRulesDir(cwd)
	if err != nil {
		return err
	}
	if created > 0 {
		fmt.Printf("✓ .agents/rules/ created (%d rule files)\n", created)
	} else {
		fmt.Println("✓ .agents/rules/ already exists (no files overwritten)")
	}

	// Step 3: Generate/update provider files
	if err := internal.InitProviderFiles(cwd); err != nil {
		return err
	}
	fmt.Println("✓ Provider files updated (CLAUDE.md, AGENTS.md, GEMINI.md)")

	// Step 4: Hint about setup if plugin roles might be needed
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run 'agent-team setup' to detect providers and install bundled roles")
	fmt.Println("  2. Create roles:   agent-team role-repo find <query>")
	fmt.Println("  3. Create workers: agent-team worker create <role>")

	return nil
}
