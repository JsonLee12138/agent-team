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
	var model string
	var newWindow bool
	cmd := &cobra.Command{
		Use:   "open <worker-id> [provider]",
		Short: "Open a worker session in a new terminal tab or window",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := ""
			if len(args) > 1 {
				provider = args[1]
			}
			return GetApp(cmd).RunWorkerOpen(args[0], provider, model, newWindow)
		},
	}
	cmd.Flags().StringVarP(&model, "model", "m", "", "AI model identifier")
	cmd.Flags().BoolVarP(&newWindow, "new-window", "w", false, "Open in a new window instead of a tab")
	return cmd
}

func (a *App) RunWorkerOpen(workerID, provider, model string, newWindow bool) error {
	root := a.Git.Root()
	configPath := internal.WorkerConfigPath(root, workerID)
	wtPath := internal.WtPath(root, a.WtBase, workerID)

	cfg, err := internal.LoadWorkerConfig(configPath)
	if err != nil {
		return fmt.Errorf("worker '%s' not found: %w", workerID, err)
	}

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("worker worktree '%s' not found at %s", workerID, wtPath)
	}

	if provider == "" {
		provider = cfg.DefaultProvider
		if provider == "" {
			provider = "claude"
		}
	}

	if a.Session.PaneAlive(cfg.PaneID) {
		fmt.Printf("Worker '%s' is already running (pane %s)\n", workerID, cfg.PaneID)
		return nil
	}

	// Copy skills to worktree
	fmt.Printf("  Copying skills for role '%s'...\n", cfg.Role)
	if err := internal.CopySkillsToWorktree(wtPath, root, cfg.Role); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to copy skills: %v\n", err)
	}

	// Inject role prompt into CLAUDE.md and AGENTS.md
	if err := internal.InjectRolePrompt(wtPath, workerID, cfg.Role, root); err != nil {
		return fmt.Errorf("inject role prompt: %w", err)
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
	cfg.Save(configPath)

	// Wait for shell init, then launch AI
	fmt.Println("  Waiting for shell to initialize...")
	time.Sleep(2 * time.Second)

	launchCmd := internal.BuildLaunchCmd(provider, model)
	a.Session.PaneSend(paneID, launchCmd)

	fmt.Printf("✓ Opened worker '%s' (role: %s, provider: %s) [pane %s]\n", workerID, cfg.Role, provider, paneID)
	return nil
}
