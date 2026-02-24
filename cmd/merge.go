// cmd/merge.go
package cmd

import (
	"fmt"
	"os"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newMergeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "merge <name>",
		Short: "Merge a role's branch into the current branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunMerge(args[0])
		},
	}
}

func (a *App) RunMerge(name string) error {
	root := a.Git.Root()
	wtPath := internal.WtPath(root, a.WtBase, name)
	branch := "team/" + name

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("role '%s' not found", name)
	}

	mainBranch, _ := a.Git.CurrentBranch()

	fmt.Printf("Merging branch '%s' into '%s'...\n", branch, mainBranch)
	msg := fmt.Sprintf("merge: integrate work from team role '%s'", name)
	if err := a.Git.Merge(branch, msg); err != nil {
		return err
	}

	fmt.Printf("✓ Merged '%s' into %s\n", name, mainBranch)
	fmt.Printf("  → Run 'agent-team delete %s' to remove the worktree when done\n", name)
	return nil
}
