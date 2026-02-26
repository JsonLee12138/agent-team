// cmd/worker_create.go
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

// defaultOpenSpecSetup is the real OpenSpec initialization function.
var defaultOpenSpecSetup = func(wtPath string) error {
	if err := internal.EnsureOpenSpec(); err != nil {
		return fmt.Errorf("install openspec: %w", err)
	}
	return internal.OpenSpecInit(wtPath)
}

// openSpecSetup can be overridden in tests to skip OpenSpec initialization.
var openSpecSetup = defaultOpenSpecSetup

func newWorkerCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <role-name>",
		Short: "Create a new worker for a role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunWorkerCreate(args[0])
		},
	}
}

func (a *App) RunWorkerCreate(roleName string) error {
	root := a.Git.Root()

	// Check role exists in agents/teams/
	roleDir := internal.RoleDir(root, roleName)
	if _, err := os.Stat(roleDir); os.IsNotExist(err) {
		// Check global skills
		home, homeErr := os.UserHomeDir()
		if homeErr == nil {
			globalSkill := fmt.Sprintf("%s/.claude/skills/%s", home, roleName)
			if _, gErr := os.Stat(globalSkill); gErr == nil {
				fmt.Printf("Role '%s' found in global skills at %s\n", roleName, globalSkill)
				fmt.Printf("Copying to %s...\n", roleDir)
				if err := internal.CopyDirPublic(globalSkill, roleDir); err != nil {
					return fmt.Errorf("copy global skill: %w", err)
				}
			} else {
				return fmt.Errorf("role '%s' not found in agents/teams/ or global skills.\nCreate it first using the role-creator skill", roleName)
			}
		} else {
			return fmt.Errorf("role '%s' not found in agents/teams/.\nCreate it first using the role-creator skill", roleName)
		}
	}

	// Compute next worker ID
	workerID := internal.NextWorkerID(root, roleName)
	wtPath := internal.WtPath(root, a.WtBase, workerID)
	branch := "team/" + workerID

	if _, err := os.Stat(wtPath); err == nil {
		return fmt.Errorf("worker '%s' already exists at %s", workerID, wtPath)
	}

	fmt.Printf("Creating worker '%s' (role: %s)...\n", workerID, roleName)

	// Create worktree
	if err := a.Git.WorktreeAdd(wtPath, branch); err != nil {
		return err
	}

	// Write .gitignore
	if err := internal.WriteWorktreeGitignore(wtPath); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}

	// Create worker config
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	cfg := &internal.WorkerConfig{
		WorkerID:        workerID,
		Role:            roleName,
		DefaultProvider: "claude",
		DefaultModel:    "",
		PaneID:          "",
		CreatedAt:       now,
	}
	configPath := internal.WorkerConfigPath(root, workerID)
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("save worker config: %w", err)
	}

	// Initialize OpenSpec in worktree
	if err := openSpecSetup(wtPath); err != nil {
		return fmt.Errorf("openspec setup: %w", err)
	}

	fmt.Printf("✓ Created worker '%s' at %s\n", workerID, wtPath)
	fmt.Printf("  → Role: %s\n", roleName)
	fmt.Printf("  → Branch: %s\n", branch)
	fmt.Printf("  → Open with: agent-team worker open %s\n", workerID)
	return nil
}
