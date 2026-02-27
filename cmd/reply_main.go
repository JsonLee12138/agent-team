// cmd/reply_main.go
package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newReplyMainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   `reply-main "<message>"`,
		Short: "Send a message to the main controller's session (used by workers)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunReplyMain(args[0])
		},
	}
}

// resolveWorktreeRoot uses git to find the actual worktree root directory,
// which is reliable even when called from a subdirectory.
// This is a variable so tests can override it.
var resolveWorktreeRoot = func() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (a *App) RunReplyMain(message string) error {
	// Use git to find the worktree root, not cwd (which could be a subdirectory)
	worktreeRoot, err := resolveWorktreeRoot()
	if err != nil {
		return fmt.Errorf("could not determine worktree root: %w", err)
	}
	workerID := filepath.Base(worktreeRoot)

	root := a.Git.Root()
	configPath := internal.WorkerConfigPath(root, workerID)
	wcfg, err := internal.LoadWorkerConfig(configPath)
	if err != nil {
		return fmt.Errorf("worker '%s' not found: %w", workerID, err)
	}

	if wcfg.ControllerPaneID == "" {
		return fmt.Errorf("no controller pane ID stored for worker '%s' — was the session opened with agent-team worker open?", workerID)
	}
	if !a.Session.PaneAlive(wcfg.ControllerPaneID) {
		return fmt.Errorf("main controller (pane %s) is not running", wcfg.ControllerPaneID)
	}

	a.Session.PaneSend(wcfg.ControllerPaneID, fmt.Sprintf("[Worker: %s] %s", workerID, message))
	fmt.Printf("✓ Sent to main controller from worker '%s'\n", workerID)
	return nil
}
