// cmd/worker_create.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

// workerShellInitDelay is overridable in tests to avoid real-time sleeps.
var workerShellInitDelay = 3 * time.Second

// defaultTaskSetup is the real task initialization function.
var defaultTaskSetup = func(_ string) error {
	return nil
}

// taskSetup can be overridden in tests to skip task initialization.
var taskSetup = defaultTaskSetup

// skillInstaller can be overridden in tests to skip npx skill installation.
var skillInstaller = internal.InstallSkillsForWorkerFromPath

func newWorkerCreateCmd() *cobra.Command {
	var provider string
	var model string
	var fresh bool
	cmd := &cobra.Command{
		Use:   "create <role-name> [--provider <provider>] [--model <model>]",
		Short: "Create a worker stub (task assign is preferred)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunWorkerCreate(args[0], provider, model, fresh)
		},
	}
	cmd.Flags().StringVarP(&provider, "provider", "p", "", workerProviderFlagHelp)
	cmd.Flags().StringVarP(&model, "model", "m", "", "AI model identifier")
	cmd.Flags().BoolVar(&fresh, "fresh", false, "Force re-install all skills, ignoring project cache")
	return cmd
}

func (a *App) RunWorkerCreate(roleName, provider, model string, fresh bool) error {
	_ = fresh
	root := a.Git.Root()

	projectRulesDir := filepath.Join(internal.ResolveAgentsDir(root), "rules", "project")
	if info, err := os.Stat(projectRulesDir); err != nil || !info.IsDir() {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Warning: project rules are missing. Run 'agent-team rules sync' to generate .agent-team/rules/project/ before assigning command-heavy work.")
		} else if err == nil {
			fmt.Fprintf(os.Stderr, "Warning: %s exists but is not a directory\n", projectRulesDir)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: could not inspect %s: %v\n", projectRulesDir, err)
		}
	}

	// 1. Resolve role (project → global priority)
	match, err := internal.ResolveRole(root, roleName)
	if err != nil {
		return err
	}

	rolePath := match.Path
	roleScope := match.Scope

	if match.Scope == "global" {
		fmt.Printf("Role '%s' found in global roles: %s\n", roleName, match.Path)
		if match.Description != "" {
			fmt.Printf("  Description: %s\n", match.Description)
		}
		fmt.Print("Use this global role? [Y/n]: ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer == "n" || answer == "no" {
			return fmt.Errorf("aborted: global role '%s' not confirmed", roleName)
		}
	}

	// 2. Determine provider
	if provider == "" {
		provider = "claude"
	}
	if err := validateWorkerProvider(provider); err != nil {
		return err
	}

	// 3. Compute next worker ID
	workerID := internal.NextWorkerID(root, a.WtBase, roleName)
	wtPath := internal.WtPath(root, a.WtBase, workerID)
	branch := "team/" + workerID
	configPath := internal.WorkerConfigPath(root, workerID)

	if _, err := os.Stat(wtPath); err == nil {
		return fmt.Errorf("worker '%s' already exists at %s", workerID, wtPath)
	}
	if a.Git.BranchExists(branch) {
		return fmt.Errorf("worker '%s' already exists on branch %s", workerID, branch)
	}

	fmt.Printf("Creating worker '%s' (role: %s, provider: %s)...\n", workerID, roleName, provider)

	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	worktreeCreated := true
	cfg := &internal.WorkerConfig{
		WorkerID:        workerID,
		Role:            roleName,
		Provider:       provider,
		DefaultModel:   model,
		PaneID:         "",
		CreatedAt:      now,
		WorktreeCreated: &worktreeCreated,
	}
	if roleScope == "global" {
		cfg.RoleScope = roleScope
		cfg.RolePath = rolePath
	}

	if err := a.Git.WorktreeAdd(wtPath, branch); err != nil {
		return err
	}
	if err := internal.WriteWorktreeGitignore(wtPath); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("save worker config: %w", err)
	}

	fmt.Printf("✓ Created worker '%s'\n", workerID)
	fmt.Printf("  → Role: %s\n", roleName)
	if roleScope == "global" {
		fmt.Printf("  → Role source: global (%s)\n", rolePath)
	}
	fmt.Printf("  → Provider: %s\n", provider)
	fmt.Printf("  → Branch: %s\n", branch)
	fmt.Printf("  → Worktree: %s\n", wtPath)
	fmt.Printf("  → Preferred flow: create a task, then run 'agent-team task assign <task-id>'\n")
	fmt.Printf("  → Compatibility flow: run 'agent-team worker open %s' to start this worker manually\n", workerID)
	return nil
}
