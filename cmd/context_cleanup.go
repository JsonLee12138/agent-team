package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newContextCleanupCmd() *cobra.Command {
	var workerID string

	cmd := &cobra.Command{
		Use:   "context-cleanup",
		Short: "Show the index-first recovery anchors for controller or worker context cleanup",
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunContextCleanup(workerID)
		},
	}

	cmd.Flags().StringVar(&workerID, "worker", "", "Worker ID to inspect from repo root")
	return cmd
}

func (a *App) RunContextCleanup(workerID string) error {
	currentRoot, err := resolveGitTopLevel()
	if err != nil {
		return fmt.Errorf("could not determine current git root: %w", err)
	}

	if !samePath(currentRoot, a.Git.Root()) {
		return a.runWorkerContextCleanup(currentRoot)
	}
	if strings.TrimSpace(workerID) != "" {
		return a.runControllerContextCleanupForWorker(strings.TrimSpace(workerID))
	}
	return a.runControllerContextCleanup()
}

func (a *App) runControllerContextCleanup() error {
	rulesIndex := filepath.Join(a.Git.Root(), ".agent-team", "rules", "index.md")
	fmt.Printf("context-cleanup (controller)\n")
	fmt.Printf("1. Reset the session context.\n")
	fmt.Printf("2. Read %s first.\n", rulesIndex)
	fmt.Printf("3. Open only the matching rule files.\n")
	fmt.Printf("4. Then reopen the current workflow/task artifacts you actually need.\n")
	fmt.Printf("5. Keep this as session cleanup plus file re-anchoring, not context compression, and do not bulk-scan all context files.\n")
	return nil
}

func (a *App) runControllerContextCleanupForWorker(workerID string) error {
	cfg, _, err := internal.LoadWorkerConfigByID(a.Git.Root(), a.WtBase, workerID)
	if err != nil {
		return fmt.Errorf("worker '%s' not found: %w", workerID, err)
	}
	fmt.Printf("context-cleanup (worker %s from controller)\n", workerID)
	fmt.Printf("1. Reset the session context.\n")
	fmt.Printf("2. Read %s first.\n", internal.WorkerConfigWritePath(a.Git.Root(), a.WtBase, workerID))
	if strings.TrimSpace(cfg.TaskID) != "" {
		fmt.Printf("3. Then read %s.\n", internal.TaskYAMLPath(a.Git.Root(), cfg.TaskID))
		fmt.Printf("4. Read %s only if needed after task.yaml.\n", internal.TaskContextPath(a.Git.Root(), cfg.TaskID))
	} else if strings.TrimSpace(cfg.TaskPath) != "" {
		taskPath := filepath.Join(a.Git.Root(), filepath.FromSlash(cfg.TaskPath))
		fmt.Printf("3. Then read %s/task.yaml if it exists.\n", taskPath)
		fmt.Printf("4. Read %s/context.md only if needed after task.yaml.\n", taskPath)
	} else {
		fmt.Printf("3. No task is currently bound; stop after worker.yaml unless a new assignment names more files.\n")
	}
	fmt.Printf("5. Keep this as session cleanup plus file re-anchoring, not context compression, and do not bulk-scan all context files.\n")
	return nil
}

func (a *App) runWorkerContextCleanup(worktreeRoot string) error {
	projectRoot, err := internal.ResolveProjectRootFromWorktree(worktreeRoot)
	if err != nil {
		return fmt.Errorf("could not determine project root: %w", err)
	}
	workerID := filepath.Base(worktreeRoot)
	cfg, _, err := internal.LoadWorkerConfigByID(projectRoot, internal.FindWtBase(projectRoot), workerID)
	if err != nil {
		return fmt.Errorf("worker '%s' not found: %w", workerID, err)
	}

	fmt.Printf("context-cleanup (worker %s)\n", workerID)
	fmt.Printf("1. Reset the session context.\n")
	fmt.Printf("2. Read %s first.\n", internal.WorkerYAMLPath(worktreeRoot))
	if strings.TrimSpace(cfg.TaskID) != "" {
		fmt.Printf("3. Then read %s.\n", internal.TaskYAMLPath(projectRoot, cfg.TaskID))
		fmt.Printf("4. Read %s only if needed after task.yaml.\n", internal.TaskContextPath(projectRoot, cfg.TaskID))
	} else if strings.TrimSpace(cfg.TaskPath) != "" {
		taskPath := filepath.Join(projectRoot, filepath.FromSlash(cfg.TaskPath))
		fmt.Printf("3. Then read %s/task.yaml if it exists.\n", taskPath)
		fmt.Printf("4. Read %s/context.md only if needed after task.yaml.\n", taskPath)
	} else {
		fmt.Printf("3. No task is currently bound; stop after worker.yaml unless the controller names more files.\n")
	}
	fmt.Printf("5. Keep this as session cleanup plus file re-anchoring, not context compression, and do not bulk-scan all context files.\n")
	return nil
}
