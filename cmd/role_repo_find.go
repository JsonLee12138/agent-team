package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/cobra"
)

func newRoleRepoFindCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "find [query]",
		Short: "Find role repositories on GitHub",
		Long:  "Search GitHub for role repositories. With a query argument, shows results directly. Without arguments in interactive mode, prompts for input and offers to install.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var query string
			if len(args) > 0 {
				query = args[0]
			}
			return GetApp(cmd).RunRoleRepoFind(cmd.InOrStdin(), cmd.OutOrStdout(), query)
		},
	}
}

func (a *App) RunRoleRepoFind(in io.Reader, out io.Writer, query string) error {
	// Direct mode: query provided as argument
	if query != "" {
		return a.roleRepoFindDirect(out, query)
	}

	// Interactive mode: prompt for query
	if !isInteractiveInput(in) {
		fmt.Fprintln(out, "Usage: agent-team role-repo find <query>")
		fmt.Fprintln(out, "  Search GitHub for role repositories matching the query.")
		fmt.Fprintln(out, "  Run without arguments in a terminal for interactive mode.")
		return nil
	}

	return a.roleRepoFindInteractive(in, out)
}

func (a *App) roleRepoFindDirect(out io.Writer, query string) error {
	client := newRoleRepoSearchClient()
	results, err := client.SearchRoleRepos(context.Background(), query)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		fmt.Fprintln(out, "No roles found matching your query.")
		return nil
	}

	// Limit to 6 results
	if len(results) > 6 {
		results = results[:6]
	}

	fmt.Fprintf(out, "Found %d role(s):\n\n", len(results))
	fmt.Fprintf(out, "  %-18s %-28s %s\n", "ROLE", "REPOSITORY", "PATH")
	fmt.Fprintf(out, "  %-18s %-28s %s\n", "──────────────────", "────────────────────────────", "────────────────────────")
	for _, r := range results {
		fmt.Fprintf(out, "  %-18s %-28s %s\n", r.Name, r.Repo, r.RolePath)
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Install with: agent-team role-repo add <owner/repo>")
	return nil
}

func (a *App) roleRepoFindInteractive(in io.Reader, out io.Writer) error {
	// Prompt for query
	var query string
	prompt := &survey.Input{
		Message: "Search roles:",
	}
	if err := survey.AskOne(prompt, &query, survey.WithValidator(survey.Required)); err != nil {
		if err == terminal.InterruptErr {
			return nil
		}
		return err
	}

	client := newRoleRepoSearchClient()
	results, err := client.SearchRoleRepos(context.Background(), query)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		fmt.Fprintln(out, "No roles found matching your query.")
		return nil
	}

	// Limit to 6 results
	if len(results) > 6 {
		results = results[:6]
	}

	// Build numbered options
	options := make([]string, len(results))
	for i, r := range results {
		options[i] = fmt.Sprintf("%s (%s)", r.Name, r.Repo)
	}

	// Prompt to select
	selected, err := promptSingleChoice(in, out, "Select a role to install:", options, options[0])
	if err != nil {
		return err
	}

	// Find the matching result
	for i, opt := range options {
		if opt == selected {
			r := results[i]
			fmt.Fprintf(out, "\nInstalling from %s...\n", r.Repo)
			return a.RunRoleRepoAdd(in, out, r.Repo, []string{r.Name}, false, false, false, false)
		}
	}

	return nil
}
