package cmd

import (
	"fmt"
	"os"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

var rebuildProjectRules = internal.RebuildProjectRules
var rebuildProjectCommands = internal.RebuildProjectCommands

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize agent-team for this project",
		Long: `Creates project-level structure under .agent-team/, including .agent-team/teams/ and .agent-team/rules/.
It writes the fixed rules entry/core files, regenerates AI-based project rule files under
.agent-team/rules/project/, and updates root provider files (CLAUDE.md, AGENTS.md, GEMINI.md).

Built-in entry/core rule files are created if missing. Project rules are regenerated on each run.
Provider files only update the tagged section.

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

	// Step 1: Initialize project structure (.agent-team/teams/)
	if err := internal.InitProject(cwd); err != nil {
		return err
	}
	fmt.Println("✓ .agent-team/teams/ ready")

	// Step 2: Initialize fixed rules entry/core files
	created, err := internal.InitRulesDirV2(cwd)
	if err != nil {
		return err
	}
	if created > 0 {
		fmt.Printf("✓ .agent-team/rules/ ready (%d entry/core rule files created)\n", created)
	} else {
		fmt.Println("✓ .agent-team/rules/ already exists (entry/core files preserved)")
	}

	// Step 3: Generate/update project rules
	scan, err := rebuildProjectRules(cwd)
	if err != nil {
		return err
	}
	if len(scan.Files) > 0 {
		fmt.Printf("✓ project rules regenerated from %d detected command file(s)\n", len(scan.Files))
	} else {
		fmt.Println("✓ project rules regenerated (no known command files detected)")
	}

	// Step 4: Generate/update provider files
	if err := internal.InitProviderFiles(cwd); err != nil {
		return err
	}
	fmt.Println("✓ Provider files updated (CLAUDE.md, AGENTS.md, GEMINI.md)")

	// Step 5: Validate generated rules after writing
	if err := internal.ValidateRules(cwd); err != nil {
		return err
	}
	fmt.Println("✓ Rules validation passed")

	// Step 5: Hint about setup if plugin roles might be needed
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run 'agent-team setup' to detect providers and install bundled roles")
	fmt.Println("  2. Create roles:   agent-team role-repo find <query>")
	fmt.Println("  3. Create workers: agent-team worker create <role>")

	return nil
}
