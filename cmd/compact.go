package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newCompactCmd() *cobra.Command {
	var paneID string
	var workerID string
	var target string
	var message string

	cmd := &cobra.Command{
		Use:   "compact",
		Short: "Send /compact to a recorded Claude session pane",
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunCompact(paneID, workerID, target, message)
		},
	}

	cmd.Flags().StringVar(&paneID, "pane-id", "", "Direct pane ID override")
	cmd.Flags().StringVar(&workerID, "worker", "", "Worker ID to target from repo root")
	cmd.Flags().StringVar(&target, "to", "", "Target alias: main")
	cmd.Flags().StringVar(&message, "message", "", "Optional /compact message")
	return cmd
}

var resolveGitTopLevel = func() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (a *App) RunCompact(paneID, workerID, target, message string) error {
	if target != "" && target != "main" {
		return fmt.Errorf("unsupported --to value %q (only 'main' is supported)", target)
	}

	text := "/compact"
	if strings.TrimSpace(message) != "" {
		text += " " + strings.TrimSpace(message)
	}

	resolvedPaneID, _, err := a.resolveCompactTarget(paneID, workerID, target)
	if err != nil {
		return err
	}
	if !a.Session.PaneAlive(resolvedPaneID) {
		return fmt.Errorf("target pane %s is not running", resolvedPaneID)
	}
	if err := a.Session.PaneSend(resolvedPaneID, text); err != nil {
		return fmt.Errorf("send /compact to pane %s: %w", resolvedPaneID, err)
	}

	fmt.Printf("✓ Sent /compact to pane %s\n", resolvedPaneID)
	return nil
}

func (a *App) resolveCompactTarget(explicitPaneID, workerID, target string) (paneID string, backend string, err error) {
	if explicitPaneID != "" {
		return explicitPaneID, "", nil
	}

	root := a.Git.Root()
	currentRoot, err := resolveGitTopLevel()
	if err != nil {
		return "", "", fmt.Errorf("could not determine current git root: %w", err)
	}
	isWorkerWorktree := !samePath(currentRoot, root)

	if target == "main" {
		if isWorkerWorktree {
			workerID := filepath.Base(currentRoot)
			cfg, _, err := internal.LoadWorkerConfigByID(root, a.WtBase, workerID)
			if err != nil {
				return "", "", fmt.Errorf("load worker config for current worktree: %w", err)
			}
			if cfg.ControllerPaneID == "" {
				return "", "", fmt.Errorf("no controller pane ID stored for worker '%s'", cfg.WorkerID)
			}
			return cfg.ControllerPaneID, cfg.Provider, nil
		}
		return a.resolveProjectMainPane(root)
	}

	if isWorkerWorktree {
		workerID := filepath.Base(currentRoot)
		cfg, _, err := internal.LoadWorkerConfigByID(root, a.WtBase, workerID)
		if err != nil {
			return "", "", fmt.Errorf("load worker config for current worktree: %w", err)
		}
		if cfg.PaneID == "" {
			return "", "", fmt.Errorf("no pane ID stored for worker '%s'", cfg.WorkerID)
		}
		return cfg.PaneID, cfg.Provider, nil
	}

	if workerID == "" {
		return "", "", fmt.Errorf("must specify --worker <worker-id> when running from repo root, or use --to main")
	}

	cfg, _, err := internal.LoadWorkerConfigByID(root, a.WtBase, workerID)
	if err != nil {
		return "", "", fmt.Errorf("worker '%s' not found: %w", workerID, err)
	}
	if cfg.PaneID == "" {
		return "", "", fmt.Errorf("no pane ID stored for worker '%s'", workerID)
	}
	return cfg.PaneID, cfg.Provider, nil
}

func (a *App) resolveProjectMainPane(root string) (paneID string, backend string, err error) {
	configPath := internal.MainSessionYAMLPath(root)
	cfg, err := internal.LoadMainSessionConfig(configPath)
	if err == nil && cfg.PaneID != "" {
		return cfg.PaneID, cfg.Backend, nil
	}
	return a.fallbackProjectMainPane(root)
}

func (a *App) fallbackProjectMainPane(root string) (paneID string, backend string, err error) {
	paneID, backend = detectCurrentPaneFromEnv()
	if paneID == "" {
		return "", "", fmt.Errorf("no project main pane recorded, and neither WEZTERM_PANE nor TMUX_PANE is set")
	}
	cfg := &internal.MainSessionConfig{
		PaneID:    paneID,
		Backend:   backend,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := cfg.Save(internal.MainSessionYAMLPath(root)); err != nil {
		return "", "", fmt.Errorf("save project main pane config: %w", err)
	}
	return paneID, backend, nil
}

func detectCurrentPaneFromEnv() (paneID string, backend string) {
	if pane := strings.TrimSpace(os.Getenv("WEZTERM_PANE")); pane != "" {
		return pane, "wezterm"
	}
	if pane := strings.TrimSpace(os.Getenv("TMUX_PANE")); pane != "" {
		return pane, "tmux"
	}
	return "", ""
}

func samePath(a, b string) bool {
	aEval, errA := filepath.EvalSymlinks(a)
	bEval, errB := filepath.EvalSymlinks(b)
	if errA == nil {
		a = aEval
	}
	if errB == nil {
		b = bEval
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

