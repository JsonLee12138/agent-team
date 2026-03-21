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
	cmd.AddCommand(newRulesValidateCmd())
	return cmd
}

func newRulesSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync .agent-team/rules/ into root provider files",
		Long: `Re-injects the rules reference section into CLAUDE.md, AGENTS.md, and GEMINI.md
based on the current .agent-team/rules/ content, refreshes the fixed entry/core rule files,
and regenerates project rules under .agent-team/rules/project/. Only the tagged section is updated
in provider files; user-written content is preserved outside the tagged section.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := GetApp(cmd)
			root := app.Git.Root()

			if !internal.HasRulesDir(root) {
				return fmt.Errorf(".agent-team/rules/ not found. Run 'agent-team init' first")
			}

			written, err := internal.SyncRulesDirV2(root)
			if err != nil {
				return err
			}
			fmt.Printf("✓ Entry/core rules synced (%d files)\n", written)

			scan, err := rebuildProjectRules(root)
			if err != nil {
				return err
			}
			if len(scan.Files) > 0 {
				fmt.Printf("✓ project rules regenerated from %d detected command file(s)\n", len(scan.Files))
			} else {
				fmt.Println("✓ project rules regenerated (no known command files detected)")
			}

			if err := internal.InitProviderFiles(root); err != nil {
				return err
			}
			fmt.Println("✓ Provider files synced (CLAUDE.md, AGENTS.md, GEMINI.md)")
			if err := internal.ValidateRules(root); err != nil {
				return err
			}
			fmt.Println("✓ Rules validation passed")
			return nil
		},
	}
	return cmd
}

func newRulesValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate .agent-team/rules/ structure and scope",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := GetApp(cmd).Git.Root()
			if err := internal.ValidateRules(root); err != nil {
				return err
			}
			fmt.Println("✓ Rules validation passed")
			return nil
		},
	}
}
