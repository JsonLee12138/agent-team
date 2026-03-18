package cmd

import (
	"fmt"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newRulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Manage agent-team rules",
	}
	cmd.AddCommand(newRulesSyncCmd())
	return cmd
}

func newRulesSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync .agents/rules/ references into root provider files",
		Long: `Re-injects the rules reference section into CLAUDE.md, AGENTS.md, and GEMINI.md
based on the current .agents/rules/ content. Only the tagged section is updated;
user-written content is preserved.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := GetApp(cmd)
			root := app.Git.Root()

			if !internal.HasRulesDir(root) {
				return fmt.Errorf(".agents/rules/ not found. Run 'agent-team init' first")
			}

			if err := internal.InitProviderFiles(root); err != nil {
				return err
			}
			fmt.Println("✓ Provider files synced (CLAUDE.md, AGENTS.md, GEMINI.md)")
			return nil
		},
	}
}
