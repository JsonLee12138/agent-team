package cmd

import (
	"fmt"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

var syncRulesDir = internal.SyncRulesDir

func newRulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Manage agent-team rules",
	}
	cmd.AddCommand(newRulesSyncCmd())
	return cmd
}

func newRulesSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync .agents/rules/ references into root provider files",
		Long: `Re-injects the rules reference section into CLAUDE.md, AGENTS.md, and GEMINI.md
based on the current .agents/rules/ content, refreshes the built-in rule files,
and regenerates project-commands.md. Only the tagged section is updated in provider files;
user-written content is preserved outside the tagged section.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := GetApp(cmd)
			root := app.Git.Root()

			if !internal.HasRulesDir(root) {
				return fmt.Errorf(".agents/rules/ not found. Run 'agent-team init' first")
			}

			written, err := syncRulesDir(root)
			if err != nil {
				return err
			}
			fmt.Printf("✓ Built-in rules synced (%d files)\n", written)

			scan, err := rebuildProjectCommands(root)
			if err != nil {
				return err
			}
			if len(scan.Files) > 0 {
				fmt.Printf("✓ %s regenerated from %d detected command file(s)\n", "project-commands.md", len(scan.Files))
			} else {
				fmt.Printf("✓ %s regenerated (no known command files detected)\n", "project-commands.md")
			}

			if err := internal.InitProviderFiles(root); err != nil {
				return err
			}
			fmt.Println("✓ Provider files synced (CLAUDE.md, AGENTS.md, GEMINI.md)")
			return nil
		},
	}
	return cmd
}
