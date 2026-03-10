// cmd/hook_stop.go
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newHookStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Handle stop/session-end event (auto-commit, archive, notify)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHookStop(cmd)
		},
	}
}

func runHookStop(cmd *cobra.Command) error {
	input, err := internal.ParseHookInput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] stop: parse input: %v\n", err)
		return nil // hook should not block session exit
	}

	wt, err := internal.ResolveWorktree(input.CWD)
	if err != nil || wt == nil {
		return nil // not in an agent-team worktree, exit silently
	}

	// 1. Auto-commit uncommitted tracked changes
	autoCommit(wt.WtPath, wt.WorkerID)

	// 2. Archive all active changes
	active, err := internal.ListActiveChanges(wt.WtPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] stop: list changes: %v\n", err)
		return nil
	}
	if len(active) == 0 {
		return nil
	}

	archived := 0
	for _, change := range active {
		if err := internal.ApplyChangeTransition(change, internal.ChangeStatusArchived); err != nil {
			fmt.Fprintf(os.Stderr, "[agent-team] stop: archive '%s' skipped (%s → archived not allowed): %v\n",
				change.Name, change.Status, err)
			continue
		}
		if err := internal.SaveChange(wt.WtPath, change); err != nil {
			fmt.Fprintf(os.Stderr, "[agent-team] stop: save '%s' failed: %v\n", change.Name, err)
			continue
		}
		archived++
		fmt.Fprintf(os.Stderr, "[agent-team] stop: archived change '%s'\n", change.Name)
	}

	// 3. Notify main controller
	if archived > 0 {
		notifyMain(wt, fmt.Sprintf(
			"Session ended: %d change(s) auto-archived by worker '%s'",
			archived, wt.WorkerID))
	}

	return nil
}

// autoCommit stages and commits tracked file modifications in the worktree.
// Only stages tracked files (git add -u), does NOT add untracked files.
func autoCommit(wtPath, workerID string) {
	statusOut, err := gitExec(wtPath, "status", "--porcelain")
	if err != nil || len(strings.TrimSpace(statusOut)) == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "[agent-team] stop: uncommitted changes detected, auto-committing...\n")

	if _, err := gitExec(wtPath, "add", "-u"); err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] stop: git add -u failed: %v\n", err)
		return
	}

	// git diff --cached --quiet exits 0 when nothing is staged
	if _, err := gitExec(wtPath, "diff", "--cached", "--quiet"); err == nil {
		return
	}

	commitMsg := fmt.Sprintf("auto-commit: worker '%s' session ended with uncommitted changes", workerID)
	if _, err := gitExec(wtPath, "commit", "-m", commitMsg); err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] stop: git commit failed: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "[agent-team] stop: auto-committed changes\n")
}

// notifyMain sends a message to the main controller via reply-main command.
func notifyMain(wt *internal.WorktreeInfo, msg string) {
	notifyCmd := exec.Command("agent-team", "reply-main", msg)
	notifyCmd.Dir = wt.WtPath
	notifyCmd.Stderr = os.Stderr
	if err := notifyCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] stop: notify main failed: %v\n", err)
	}
}

// gitExec runs a git command in the given directory and returns stdout.
func gitExec(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}
