package cmd

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newRoleRepoUpdateCmd() *cobra.Command {
	var global bool
	var yes bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update installed roles from remote sources",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunRoleRepoUpdate(cmd.InOrStdin(), cmd.OutOrStdout(), global, yes)
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Update global role installs")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Overwrite existing role directories during update")
	return cmd
}

func (a *App) RunRoleRepoUpdate(in io.Reader, out io.Writer, global bool, yes bool) error {
	root := a.Git.Root()
	scope := roleRepoScopeFromFlag(global)
	installRoot, err := internal.ResolveRoleRepoInstallRoot(root, scope)
	if err != nil {
		return err
	}
	lockPath, lock, warning, err := roleRepoLockForScope(root, scope)
	if err != nil {
		return err
	}
	printRoleRepoLockWarning(warning)

	client := internal.NewRoleRepoGitHubClient()
	statuses, _ := internal.CheckRoleRepoUpdates(context.Background(), client, installRoot, lock)
	candidates := make([]string, 0)
	for _, st := range statuses {
		if st.State == "update_available" {
			candidates = append(candidates, st.Name)
		}
	}

	selected := candidates
	overwrite := true
	if !yes {
		if len(candidates) == 0 {
			fmt.Fprintln(out, "No updates available.")
			return nil
		}
		chosen, selectErr := promptSelectNames(in, out, "Select role(s) to update:", candidates)
		if selectErr != nil {
			return selectErr
		}
		selected = chosen
		ok, confirmErr := promptConfirm(in, out, fmt.Sprintf("Update %d role(s)?", len(selected)))
		if confirmErr != nil {
			return confirmErr
		}
		if !ok {
			return fmt.Errorf("update cancelled")
		}
	}

	updated, skipped, failed := internal.UpdateRoleRepoFromLock(context.Background(), client, installRoot, &lock, overwrite, selected, time.Now)
	if err := internal.WriteRoleRepoLock(lockPath, lock); err != nil {
		return err
	}

	for _, name := range updated {
		fmt.Fprintf(out, "+ updated %s\n", name)
	}
	for _, name := range skipped {
		if yes {
			fmt.Fprintf(out, "- skipped %s (already up to date)\n", name)
		} else {
			fmt.Fprintf(out, "- skipped %s (not selected or already up to date)\n", name)
		}
	}
	for name, upErr := range failed {
		fmt.Fprintf(out, "- failed %s: %v\n", name, upErr)
	}
	fmt.Fprintf(out, "Summary: updated=%d skipped=%d failed=%d\n", len(updated), len(skipped), len(failed))
	if len(failed) > 0 {
		return fmt.Errorf("one or more updates failed")
	}
	return nil
}
