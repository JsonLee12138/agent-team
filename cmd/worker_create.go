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

// skillInstaller can be overridden in tests to skip npx skill installation.
var skillInstaller = internal.InstallSkillsForWorker

func newWorkerCreateCmd() *cobra.Command {
	var model string
	var newWindow bool
	cmd := &cobra.Command{
		Use:   "create <role-name> [provider]",
		Short: "Create a new worker and open its session",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := ""
			if len(args) > 1 {
				provider = args[1]
			}
			return GetApp(cmd).RunWorkerCreate(args[0], provider, model, newWindow)
		},
	}
	cmd.Flags().StringVarP(&model, "model", "m", "", "AI model identifier")
	cmd.Flags().BoolVarP(&newWindow, "new-window", "w", false, "Open in a new window instead of a tab")
	return cmd
}

func (a *App) RunWorkerCreate(roleName, provider, model string, newWindow bool) error {
	root := a.Git.Root()

	// 1. Validate role exists
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
				return fmt.Errorf("role '%s' not found in .agents/teams/ or global skills.\nCreate it first using the role-creator skill", roleName)
			}
		} else {
			return fmt.Errorf("role '%s' not found in .agents/teams/.\nCreate it first using the role-creator skill", roleName)
		}
	}

	// 2. Determine provider
	if provider == "" {
		provider = "claude"
	}
	if !internal.SupportedProviders[provider] {
		return fmt.Errorf("unsupported provider '%s' (supported: claude, codex, opencode)", provider)
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

	// 6. Get mainSessionID
	mainSessionID := os.Getenv("WEZTERM_PANE")
	if mainSessionID == "" {
		mainSessionID = os.Getenv("TMUX_PANE")
	}

	// 7. Create and save worker.yaml to worktree root
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	cfg := &internal.WorkerConfig{
		WorkerID:      workerID,
		Role:          roleName,
		Provider:      provider,
		DefaultModel:  model,
		MainSessionID: mainSessionID,
		PaneID:        "",
		CreatedAt:     now,
	}
	configPath := internal.WorkerYAMLPath(wtPath)
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("save worker config: %w", err)
	}

	// 8. Initialize OpenSpec in worktree
	if err := openSpecSetup(wtPath); err != nil {
		return fmt.Errorf("openspec setup: %w", err)
	}

	// 9. Inject role prompt into CLAUDE.md and AGENTS.md
	if err := internal.InjectRolePrompt(wtPath, workerID, roleName, root); err != nil {
		return fmt.Errorf("inject role prompt: %w", err)
	}

	// 10. Open window — SpawnPane
	paneID, err := a.Session.SpawnPane(wtPath, newWindow)
	if err != nil || paneID == "" {
		return fmt.Errorf("failed to open session for '%s': %w", workerID, err)
	}

	// 11. Set pane title, return focus to controller
	a.Session.SetTitle(paneID, workerID)
	if !newWindow {
		if currentPane := os.Getenv("WEZTERM_PANE"); currentPane != "" {
			a.Session.ActivatePane(currentPane)
		}
	}

	// 12. Save paneID + controllerPaneID to worker.yaml
	cfg.PaneID = paneID
	if controllerPane := os.Getenv("WEZTERM_PANE"); controllerPane != "" {
		cfg.ControllerPaneID = controllerPane
	} else if controllerPane := os.Getenv("TMUX_PANE"); controllerPane != "" {
		cfg.ControllerPaneID = controllerPane
	}
	cfg.Save(configPath)

	// 13. Install skills
	fmt.Printf("  Installing skills for role '%s'...\n", roleName)
	if err := skillInstaller(wtPath, root, roleName, provider); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: skill installation had errors: %v\n", err)
	}

	// 14. Wait for shell init
	fmt.Println("  Waiting for shell to initialize...")
	time.Sleep(2 * time.Second)

	// 15. Send AI launch command
	launchCmd := internal.BuildLaunchCmd(provider, model)
	a.Session.PaneSend(paneID, launchCmd)

	fmt.Printf("✓ Created and opened worker '%s' at %s\n", workerID, wtPath)
	fmt.Printf("  → Role: %s\n", roleName)
	fmt.Printf("  → Provider: %s\n", provider)
	fmt.Printf("  → Branch: %s\n", branch)
	fmt.Printf("  → Pane: %s\n", paneID)
	return nil
}
