package cmd

import (
	"context"
	"fmt"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newRoleRepoCheckCmd() *cobra.Command {
	var global bool
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check installed roles for remote updates",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunRoleRepoCheck(global)
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Check global role installs")
	return cmd
}

func (a *App) RunRoleRepoCheck(global bool) error {
	root := a.Git.Root()
	scope := roleRepoScopeFromFlag(global)
	installRoot, err := internal.ResolveRoleRepoInstallRoot(root, scope)
	if err != nil {
		return err
	}
	_, lock, warning, err := roleRepoLockForScope(root, scope)
	if err != nil {
		return err
	}
	printRoleRepoLockWarning(warning)

	client := internal.NewRoleRepoGitHubClient()
	statuses, untracked := internal.CheckRoleRepoUpdates(context.Background(), client, installRoot, lock)
	fmt.Print(internal.FormatRoleRepoCheckSummary(statuses, untracked))

	updates := 0
	errs := 0
	for _, st := range statuses {
		if st.State == "update_available" {
			updates++
		}
		if st.State == "error" {
			errs++
		}
	}
	fmt.Printf("Summary: updates=%d errors=%d untracked=%d\n", updates, errs, len(untracked))
	if errs > 0 {
		return fmt.Errorf("check completed with errors")
	}
	return nil
}
