package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newRoleRepoAddCmd() *cobra.Command {
	var global bool
	var roleNames []string
	var listOnly bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "add <source>",
		Short: "Install roles from a repository source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scopeExplicit := cmd.Flags().Changed("global")
			return GetApp(cmd).RunRoleRepoAdd(cmd.InOrStdin(), cmd.OutOrStdout(), args[0], roleNames, global, scopeExplicit, listOnly, yes)
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Install to global scope (~/.agents/roles)")
	cmd.Flags().StringArrayVar(&roleNames, "role", nil, "Role name to install (repeatable)")
	cmd.Flags().BoolVar(&listOnly, "list", false, "List discovered roles without installing")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Overwrite existing role directories")
	return cmd
}

func (a *App) RunRoleRepoAdd(in io.Reader, out io.Writer, sourceArg string, roleNames []string, global bool, scopeExplicit bool, listOnly bool, overwrite bool) error {
	root := a.Git.Root()

	// Scope selection: if --global not explicitly set and interactive, prompt
	if !scopeExplicit && !listOnly && isInteractiveInput(in) {
		choice, err := promptSingleChoice(in, out,
			"Install scope:",
			[]string{"Project (.agents/teams/)", "Global (~/.agents/roles/)"},
			"Project (.agents/teams/)",
		)
		if err != nil {
			return err
		}
		global = strings.HasPrefix(choice, "Global")
	}

	scope := roleRepoScopeFromFlag(global)
	installRoot, err := internal.ResolveRoleRepoInstallRoot(root, scope)
	if err != nil {
		return err
	}
	if err := internal.EnsureRoleRepoInstallRoot(installRoot); err != nil {
		return err
	}
	lockPath, lock, warning, err := roleRepoLockForScope(root, scope)
	if err != nil {
		return err
	}
	printRoleRepoLockWarning(warning)

	source, err := internal.ParseRoleRepoSource(sourceArg)
	if err != nil {
		return err
	}

	client := internal.NewRoleRepoGitHubClient()
	roles, err := client.DiscoverRemoteRoles(context.Background(), source)
	if err != nil {
		return err
	}
	if len(roles) == 0 {
		return fmt.Errorf("no valid roles found; accepted paths are skills/<role>/references/role.yaml and .agents/teams/<role>/references/role.yaml")
	}

	selected, err := internal.SelectRoleRepoRemotes(roles, roleNames)
	if err != nil {
		return err
	}
	if len(roleNames) == 0 && len(roles) > 1 && !listOnly {
		options := make([]string, 0, len(roles))
		for _, role := range roles {
			options = append(options, role.Candidate.Name)
		}
		sort.Strings(options)
		chosen, selectErr := promptSelectNames(in, out, "Multiple roles found. Choose role(s) to install:", options)
		if selectErr != nil {
			return selectErr
		}
		selected, err = internal.SelectRoleRepoRemotes(roles, chosen)
		if err != nil {
			return err
		}
	}
	if listOnly {
		for _, role := range selected {
			fmt.Fprintf(out, "%s\t%s\n", role.Candidate.Name, role.Candidate.RolePath)
		}
		return nil
	}

	// Confirmation step: show summary and prompt before install
	if !overwrite && isInteractiveInput(in) {
		scopeLabel := "Project"
		if global {
			scopeLabel = "Global"
		}
		fmt.Fprintf(out, "\nInstall summary:\n")
		fmt.Fprintf(out, "  Source: %s\n", source.FullName())
		fmt.Fprintf(out, "  Scope:  %s (%s)\n", scopeLabel, installRoot)
		fmt.Fprintf(out, "  Roles:  ")
		names := make([]string, len(selected))
		for i, r := range selected {
			names[i] = r.Candidate.Name
		}
		fmt.Fprintf(out, "%s\n\n", strings.Join(names, ", "))

		confirmed, err := promptConfirm(in, out, "Proceed with installation?")
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Fprintln(out, "Installation cancelled.")
			return nil
		}
	}

	now := time.Now().UTC()
	success := 0
	failed := 0
	for _, role := range selected {
		_, backup, installErr := internal.InstallRoleRepoRemoteRole(context.Background(), client, role, installRoot, overwrite, time.Now)
		if installErr != nil {
			if errors.Is(installErr, internal.ErrRoleRepoInstallConflict) {
				fmt.Fprintf(out, "- skipped %s (already exists, use -y to overwrite)\n", role.Candidate.Name)
				continue
			}
			fmt.Fprintf(out, "- failed %s: %v\n", role.Candidate.Name, installErr)
			failed++
			continue
		}
		entry := internal.RoleRepoLockEntry{
			Name:        role.Candidate.Name,
			Source:      source.FullName(),
			SourceType:  role.Candidate.SourceType,
			SourceURL:   role.Candidate.SourceURL,
			RolePath:    role.Candidate.RolePath,
			FolderHash:  role.FolderHash,
			InstalledAt: now,
			UpdatedAt:   now,
		}
		if existing, ok := internal.FindRoleRepoLockEntry(lock, role.Candidate.Name); ok {
			entry.InstalledAt = existing.InstalledAt
		}
		internal.UpsertRoleRepoLockEntry(&lock, entry)
		if backup != "" {
			fmt.Fprintf(out, "+ installed %s (backup: %s)\n", role.Candidate.Name, backup)
		} else {
			fmt.Fprintf(out, "+ installed %s\n", role.Candidate.Name)
		}
		success++
	}

	if err := internal.WriteRoleRepoLock(lockPath, lock); err != nil {
		return err
	}
	fmt.Fprintf(out, "Done. success=%d failed=%d\n", success, failed)
	if failed > 0 {
		return fmt.Errorf("one or more roles failed to install")
	}
	return nil
}
