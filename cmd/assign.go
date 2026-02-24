// cmd/assign.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/leeforge/agent-team/internal"
	"github.com/spf13/cobra"
)

func newAssignCmd() *cobra.Command {
	var model string
	cmd := &cobra.Command{
		Use:   `assign <name> "<task>" [provider]`,
		Short: "Write a task file and notify the role session",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := ""
			if len(args) > 2 {
				provider = args[2]
			}
			return GetApp(cmd).RunAssign(args[0], args[1], provider, model)
		},
	}
	cmd.Flags().StringVarP(&model, "model", "m", "", "AI model identifier")
	return cmd
}

func (a *App) RunAssign(name, task, provider, model string) error {
	root := a.Git.Root()
	teamsDir := internal.TeamsDir(root, a.WtBase, name)
	configPath := internal.ConfigPath(root, a.WtBase, name)

	if _, err := os.Stat(teamsDir); os.IsNotExist(err) {
		return fmt.Errorf("role '%s' not found", name)
	}

	// Create task file
	ts := time.Now().Format("2006-01-02-15-04-05")
	slug := internal.Slugify(task, 50)
	fileName := fmt.Sprintf("%s-%s.md", ts, slug)
	taskPath := filepath.Join(teamsDir, "tasks", "pending", fileName)

	nowUTC := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	content := fmt.Sprintf("# Task: %s\n\nAssigned: %s\nStatus: pending\n\n## Description\n\n%s\n\n## Notes\n\n_Add implementation notes here_\n",
		task, nowUTC, task)

	if err := os.WriteFile(taskPath, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("✓ Task file: %s\n", taskPath)

	// Ensure session is running
	cfg, err := internal.LoadRoleConfig(configPath)
	if err != nil {
		return err
	}

	if !a.Session.PaneAlive(cfg.PaneID) {
		fmt.Printf("Role '%s' is not running, opening session first...\n", name)
		if err := a.RunOpen(name, provider, model); err != nil {
			return err
		}
		// Reload config to get new pane ID
		cfg, err = internal.LoadRoleConfig(configPath)
		if err != nil {
			return err
		}
		fmt.Println("  Waiting for AI to initialize...")
		time.Sleep(3 * time.Second)
	}

	// Notify
	taskRel := fmt.Sprintf("agents/teams/%s/tasks/pending/%s", name, fileName)
	msg := fmt.Sprintf("New task assigned: %s\nPlease read the task file at: %s\nWhen complete, move it to agents/teams/%s/tasks/done/",
		task, taskRel, name)
	a.Session.PaneSend(cfg.PaneID, msg)

	fmt.Printf("✓ Assigned to '%s': %s\n", name, task)
	return nil
}
