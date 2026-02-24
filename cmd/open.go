// cmd/open.go
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newOpenCmd() *cobra.Command {
	var model string
	var newWindow bool
	cmd := &cobra.Command{
		Use:   "open <name> [provider]",
		Short: "Open a role session in a new terminal tab or window",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := ""
			if len(args) > 1 {
				provider = args[1]
			}
			return GetApp(cmd).RunOpen(args[0], provider, model, newWindow)
		},
	}
	cmd.Flags().StringVarP(&model, "model", "m", "", "AI model identifier")
	cmd.Flags().BoolVarP(&newWindow, "new-window", "w", false, "Open in a new window instead of a tab")
	return cmd
}

func newOpenAllCmd() *cobra.Command {
	var model string
	var newWindow bool
	cmd := &cobra.Command{
		Use:   "open-all [provider]",
		Short: "Open all role sessions",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := ""
			if len(args) > 0 {
				provider = args[0]
			}
			return GetApp(cmd).RunOpenAll(provider, model, newWindow)
		},
	}
	cmd.Flags().StringVarP(&model, "model", "m", "", "AI model identifier")
	cmd.Flags().BoolVarP(&newWindow, "new-window", "w", false, "Open in a new window instead of a tab")
	return cmd
}

func (a *App) RunOpen(name, provider, model string, newWindow bool) error {
	root := a.Git.Root()
	teamsDir := internal.TeamsDir(root, a.WtBase, name)
	configPath := internal.ConfigPath(root, a.WtBase, name)
	wtPath := internal.WtPath(root, a.WtBase, name)

	if _, err := os.Stat(teamsDir); os.IsNotExist(err) {
		return fmt.Errorf("role '%s' not found", name)
	}

	cfg, err := internal.LoadRoleConfig(configPath)
	if err != nil {
		return err
	}

	if provider == "" {
		provider = cfg.DefaultProvider
		if provider == "" {
			provider = "claude"
		}
	}

	if a.Session.PaneAlive(cfg.PaneID) {
		fmt.Printf("Role '%s' is already running (pane %s)\n", name, cfg.PaneID)
		return nil
	}

	// Inject role prompt into CLAUDE.md and AGENTS.md
	if err := internal.InjectRolePrompt(wtPath, name, root); err != nil {
		return fmt.Errorf("inject role prompt: %w", err)
	}

	// Spawn pane
	paneID, err := a.Session.SpawnPane(wtPath, newWindow)
	if err != nil || paneID == "" {
		return fmt.Errorf("failed to open session for '%s': %w", name, err)
	}

	a.Session.SetTitle(paneID, name)

	// Return focus (wezterm tab mode only; new window doesn't steal focus)
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

	fmt.Printf("✓ Opened role '%s' (%s) [pane %s]\n", name, provider, paneID)
	return nil
}

func (a *App) RunOpenAll(provider, model string, newWindow bool) error {
	root := a.Git.Root()
	roles := internal.ListRoles(root, a.WtBase)
	if len(roles) == 0 {
		return fmt.Errorf("no roles found. Create one with: agent-team create <name>")
	}
	for _, role := range roles {
		if err := a.RunOpen(role, provider, model, newWindow); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to open '%s': %v\n", role, err)
		}
	}
	return nil
}
