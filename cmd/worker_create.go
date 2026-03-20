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
var defaultTaskSetup = func(wtPath string) error {
	return internal.InitTasksDir(wtPath)
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
		Short: "Create a new worker (use 'worker open' to start its session)",
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

	projectCommandsPath := filepath.Join(internal.ResolveAgentsDir(root), "rules", "project-commands.md")
	if _, err := os.Stat(projectCommandsPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Warning: project command rules are missing. Run 'agent-team rules sync' to generate .agents/rules/project-commands.md before assigning command-heavy work.")
		} else {
			fmt.Fprintf(os.Stderr, "Warning: could not inspect %s: %v\n", projectCommandsPath, err)
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
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("worker '%s' already exists at %s", workerID, configPath)
	}

	fmt.Printf("Creating worker '%s' (role: %s, provider: %s)...\n", workerID, roleName, provider)

	// 4. Create and save centralized worker config only.
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	worktreeCreated := false
	cfg := &internal.WorkerConfig{
		WorkerID:        workerID,
		Role:            roleName,
		Provider:        provider,
		DefaultModel:    model,
		PaneID:          "",
		CreatedAt:       now,
		WorktreeCreated: &worktreeCreated,
	}
	if roleScope == "global" {
		cfg.RoleScope = roleScope
		cfg.RolePath = rolePath
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
	fmt.Printf("  → Worktree: deferred until first open/assign\n")
	fmt.Printf("  → Run 'agent-team worker open %s' to create the worktree and start the session\n", workerID)
	return nil
}
