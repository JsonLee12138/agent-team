// cmd/worker_open.go
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newWorkerOpenCmd() *cobra.Command {
	var provider string
	var model string
	var newWindow bool
	cmd := &cobra.Command{
		Use:   "open <worker-id> [--provider <provider>] [--model <model>] [--new-window]",
		Short: "Reopen an existing worker session in a new terminal tab or window",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			providerChanged := cmd.Flags().Changed("provider")
			modelChanged := cmd.Flags().Changed("model")
			return GetApp(cmd).RunWorkerOpen(args[0], provider, model, newWindow, providerChanged, modelChanged)
		},
	}
	cmd.Flags().StringVarP(&provider, "provider", "p", "", workerProviderFlagHelp)
	cmd.Flags().StringVarP(&model, "model", "m", "", "AI model identifier")
	cmd.Flags().BoolVarP(&newWindow, "new-window", "w", false, "Open in a new window instead of a tab")
	return cmd
}

func (a *App) bootstrapWorkerWorktree(root, wtPath, workerID, sessionProvider string, cfg *internal.WorkerConfig, fresh bool) (string, error) {
	branch := "team/" + workerID
	if err := a.Git.WorktreeAdd(wtPath, branch); err != nil {
		return "", err
	}
	if err := internal.WriteWorktreeGitignore(wtPath); err != nil {
		return "", fmt.Errorf("write .gitignore: %w", err)
	}
	rolePath := cfg.RolePath
	if rolePath == "" {
		rolePath = internal.RoleDir(root, cfg.Role)
	}
	if err := internal.InjectRolePromptWithPath(wtPath, workerID, cfg.Role, rolePath, root); err != nil {
		return "", fmt.Errorf("inject role prompt: %w", err)
	}
	if err := skillInstaller(wtPath, root, cfg.Role, rolePath, sessionProvider, fresh); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: skill sync had errors: %v\n", err)
	}
	if err := taskSetup(wtPath); err != nil {
		return "", fmt.Errorf("task setup: %w", err)
	}
	worktreeCreated := true
	cfg.WorktreeCreated = &worktreeCreated
	configPath := internal.WorkerConfigPath(root, workerID)
	if err := cfg.Save(configPath); err != nil {
		return "", fmt.Errorf("save worker config: %w", err)
	}
	return configPath, nil
}

func (a *App) RunWorkerOpen(workerID, provider, model string, newWindow, persistProvider, persistModel bool) error {
	root := a.Git.Root()
	wtPath := internal.WtPath(root, a.WtBase, workerID)

	cfg, configPath, err := internal.LoadWorkerConfigByID(root, a.WtBase, workerID)
	if err != nil {
		return fmt.Errorf("worker '%s' not found: %w", workerID, err)
	}

	if provider != "" {
		if err := validateWorkerProvider(provider); err != nil {
			return err
		}
	}

	sessionProvider := cfg.Provider
	if provider != "" {
		sessionProvider = provider
	}
	usedCompatProviderFallback := false
	if sessionProvider == "" {
		sessionProvider = "claude"
		usedCompatProviderFallback = true
	}

	sessionModel := cfg.DefaultModel
	if persistModel || model != "" {
		sessionModel = model
	}

	configChanged := false
	if persistProvider {
		cfg.Provider = provider
		configChanged = true
	}
	if persistModel {
		cfg.DefaultModel = model
		configChanged = true
	}
	if configChanged {
		if err := cfg.Save(configPath); err != nil {
			return fmt.Errorf("save worker config: %w", err)
		}
	}

	bootstrapped := false
	if !cfg.IsWorktreeCreated() {
		fmt.Printf("  Bootstrapping worktree for worker '%s'...\n", workerID)
		configPath, err = a.bootstrapWorkerWorktree(root, wtPath, workerID, sessionProvider, cfg, false)
		if err != nil {
			return err
		}
		bootstrapped = true
	} else if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("worker worktree '%s' not found at %s", workerID, wtPath)
	}

	if a.Session.PaneAlive(cfg.PaneID) {
		if configChanged {
			fmt.Printf("Worker '%s' is already running (pane %s); saved provider/model overrides for future launches\n", workerID, cfg.PaneID)
		} else {
			fmt.Printf("Worker '%s' is already running (pane %s)\n", workerID, cfg.PaneID)
		}
		return nil
	}

	if usedCompatProviderFallback {
		fmt.Println("  worker.yaml has no provider; using claude for this launch only")
	}
	if !bootstrapped {
		fmt.Printf("  Syncing skills for role '%s'...\n", cfg.Role)
		rolePath := cfg.RolePath
		if rolePath == "" {
			rolePath = internal.RoleDir(root, cfg.Role)
		}
		if err := skillInstaller(wtPath, root, cfg.Role, rolePath, sessionProvider, false); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skill sync had errors: %v\n", err)
		}
		if err := internal.InjectRolePromptWithPath(wtPath, workerID, cfg.Role, rolePath, root); err != nil {
			return fmt.Errorf("inject role prompt: %w", err)
		}
	}

	// Spawn pane
	paneID, err := a.Session.SpawnPane(wtPath, newWindow)
	if err != nil || paneID == "" {
		return fmt.Errorf("failed to open session for '%s': %w", workerID, err)
	}

	a.Session.SetTitle(paneID, workerID)

	// Return focus (wezterm tab mode only)
	if !newWindow {
		if currentPane := os.Getenv("WEZTERM_PANE"); currentPane != "" {
			a.Session.ActivatePane(currentPane)
		}
	}

	// Save pane ID and controller pane ID
	cfg.PaneID = paneID
	if controllerPane := os.Getenv("WEZTERM_PANE"); controllerPane != "" {
		cfg.ControllerPaneID = controllerPane
	} else if controllerPane := os.Getenv("TMUX_PANE"); controllerPane != "" {
		cfg.ControllerPaneID = controllerPane
	}
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("save worker config: %w", err)
	}

	// Wait for shell init, then launch AI
	fmt.Println("  Waiting for shell to initialize...")
	time.Sleep(workerShellInitDelay)

	launchCmd := internal.BuildLaunchCmd(sessionProvider, sessionModel)
	a.Session.PaneSend(paneID, launchCmd)

	fmt.Printf("✓ Opened worker '%s' (role: %s, provider: %s) [pane %s]\n", workerID, cfg.Role, sessionProvider, paneID)
	return nil
}
