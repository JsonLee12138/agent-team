package cmd

import (
	"context"
	"fmt"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newRoleRepoSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search role repositories on GitHub",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunRoleRepoSearch(args[0])
		},
	}
}

func (a *App) RunRoleRepoSearch(query string) error {
	client := internal.NewRoleRepoGitHubClient()
	results, err := client.SearchRoleRepos(context.Background(), query)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		fmt.Println("No roles found matching strict contracts.")
		return nil
	}
	fmt.Printf("%-20s %-28s %s\n", "Role", "Repository", "Role Path")
	fmt.Printf("%-20s %-28s %s\n", "────────────────────", "────────────────────────────", "────────────────────────")
	for _, r := range results {
		fmt.Printf("%-20s %-28s %s\n", r.Name, r.Repo, r.RolePath)
	}
	return nil
}
