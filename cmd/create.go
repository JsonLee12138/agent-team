// cmd/create.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new role with git worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunCreate(args[0])
		},
	}
}

func (a *App) RunCreate(name string) error {
	root := a.Git.Root()
	wtPath := internal.WtPath(root, a.WtBase, name)
	branch := "team/" + name

	if _, err := os.Stat(wtPath); err == nil {
		return fmt.Errorf("role '%s' already exists at %s", name, wtPath)
	}

	fmt.Printf("Creating role '%s'...\n", name)
	if err := a.Git.WorktreeAdd(wtPath, branch); err != nil {
		return err
	}

	teamsDir := internal.TeamsDir(root, a.WtBase, name)
	for _, sub := range []string{"tasks/pending", "tasks/done"} {
		if err := os.MkdirAll(filepath.Join(teamsDir, sub), 0755); err != nil {
			return err
		}
	}

	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	cfg := &internal.RoleConfig{
		Name:            name,
		Description:     "",
		DefaultProvider: "claude",
		DefaultModel:    "",
		CreatedAt:       now,
		PaneID:          "",
	}
	if err := cfg.Save(filepath.Join(teamsDir, "config.yaml")); err != nil {
		return err
	}

	promptPath := filepath.Join(teamsDir, "prompt.md")
	if err := os.WriteFile(promptPath, []byte(internal.PromptMDContent(name)), 0644); err != nil {
		return err
	}

	fmt.Printf("✓ Created role '%s' at %s\n", name, wtPath)
	fmt.Printf("  → Edit %s/prompt.md to define the role\n", teamsDir)
	fmt.Printf("  → Edit %s/config.yaml to set default_provider\n", teamsDir)
	return nil
}
