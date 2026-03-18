// cmd/worker_create.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
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
	root := a.Git.Root()

	if needsRebuild, currentHash, storedHash, err := internal.BuildRulesNeedRebuild(root); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not verify build rule hash: %v\n", err)
	} else if needsRebuild {
		if storedHash == "" {
			fmt.Fprintf(os.Stderr, "Warning: build rules hash is missing (current: %s). Run 'agent-team rules sync --rebuild' to refresh dynamic build rules.\n", currentHash)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: build scripts changed since the last rules rebuild (stored: %s, current: %s). Run 'agent-team rules sync --rebuild' to refresh dynamic build rules.\n", storedHash, currentHash)
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

	if _, err := os.Stat(wtPath); err == nil {
		return fmt.Errorf("worker '%s' already exists at %s", workerID, wtPath)
	}

	fmt.Printf("Creating worker '%s' (role: %s, provider: %s)...\n", workerID, roleName, provider)

	// 4. Create worktree
	if err := a.Git.WorktreeAdd(wtPath, branch); err != nil {
		return err
	}

	// 5. Write .gitignore (includes worker.yaml)
	if err := internal.WriteWorktreeGitignore(wtPath); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}

	// 6. Create and save worker.yaml to worktree root
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	cfg := &internal.WorkerConfig{
		WorkerID:     workerID,
		Role:         roleName,
		Provider:     provider,
		DefaultModel: model,
		PaneID:       "",
		CreatedAt:    now,
	}
	if roleScope == "global" {
		cfg.RoleScope = roleScope
		cfg.RolePath = rolePath
	}
	configPath := internal.WorkerYAMLPath(wtPath)
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("save worker config: %w", err)
	}

	// 7. Initialize tasks in worktree
	if err := taskSetup(wtPath); err != nil {
		return fmt.Errorf("task setup: %w", err)
	}

	// 8. Inject role prompt into CLAUDE.md and AGENTS.md
	if err := internal.InjectRolePromptWithPath(wtPath, workerID, roleName, rolePath, root); err != nil {
		return fmt.Errorf("inject role prompt: %w", err)
	}

	// 9. Install skills
	fmt.Printf("  Installing skills for role '%s'...\n", roleName)
	if err := skillInstaller(wtPath, root, roleName, rolePath, provider, fresh); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: skill installation had errors: %v\n", err)
	}

	fmt.Printf("✓ Created worker '%s' at %s\n", workerID, wtPath)
	fmt.Printf("  → Role: %s\n", roleName)
	if roleScope == "global" {
		fmt.Printf("  → Role source: global (%s)\n", rolePath)
	}
	fmt.Printf("  → Provider: %s\n", provider)
	fmt.Printf("  → Branch: %s\n", branch)
	fmt.Printf("  → Run 'agent-team worker open %s' to start the session\n", workerID)
	return nil
}
