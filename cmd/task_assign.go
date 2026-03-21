package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newTaskAssignCmd() *cobra.Command {
	var provider string
	var model string
	var workerID string
	var newWindow bool
	cmd := &cobra.Command{
		Use:   "assign <task-id>",
		Short: "Assign a task and open its worker session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunTaskAssign(args[0], workerID, provider, model, newWindow)
		},
	}
	cmd.Flags().StringVar(&workerID, "worker", "", "Existing worker ID for same-role reassignment")
	cmd.Flags().StringVarP(&provider, "provider", "p", "", workerProviderFlagHelp)
	cmd.Flags().StringVarP(&model, "model", "m", "", "AI model identifier")
	cmd.Flags().BoolVarP(&newWindow, "new-window", "w", false, "Open in a new window instead of a tab")
	return cmd
}

func (a *App) RunTaskAssign(taskID, requestedWorkerID, provider, model string, newWindow bool) error {
	root := a.Git.Root()
	record, archived, err := internal.LoadTaskRecord(root, taskID)
	if err != nil {
		return err
	}
	if archived {
		return fmt.Errorf("task '%s' is archived", taskID)
	}
	if record.Status != internal.TaskStatusDraft && record.Status != internal.TaskStatusAssigned && record.Status != internal.TaskStatusDone {
		return fmt.Errorf("task '%s' cannot be assigned from status '%s'", taskID, record.Status)
	}
	if provider != "" {
		if err := validateWorkerProvider(provider); err != nil {
			return err
		}
	}

	workerID := requestedWorkerID
	var cfg *internal.WorkerConfig
	if workerID == "" {
		if record.Status == internal.TaskStatusAssigned && record.WorkerID != "" {
			workerID = record.WorkerID
		} else {
			workerID = internal.NextWorkerID(root, a.WtBase, record.Role)
			now := time.Now().UTC().Format(time.RFC3339)
			worktreeCreated := false
			cfg = &internal.WorkerConfig{
				WorkerID:        workerID,
				Role:            record.Role,
				Provider:        defaultWorkerProvider(provider),
				DefaultModel:    model,
				TaskID:          record.TaskID,
				TaskPath:        record.TaskPath,
				Status:          internal.TaskStatusAssigned,
				CreatedAt:       now,
				UpdatedAt:       now,
				WorktreeCreated: &worktreeCreated,
			}
			if err := cfg.Save(internal.WorkerConfigPath(root, workerID)); err != nil {
				return fmt.Errorf("save worker config: %w", err)
			}
		}
	}

	if cfg == nil {
		cfg, _, err = internal.LoadWorkerConfigByID(root, a.WtBase, workerID)
		if err != nil {
			return fmt.Errorf("load worker config: %w", err)
		}
	}
	if cfg.Role != record.Role {
		return fmt.Errorf("worker '%s' role mismatch: task requires '%s', worker has '%s'", workerID, record.Role, cfg.Role)
	}
	if provider != "" {
		cfg.Provider = provider
	}
	if model != "" {
		cfg.DefaultModel = model
	}
	cfg.TaskID = record.TaskID
	cfg.TaskPath = record.TaskPath
	cfg.Status = internal.TaskStatusAssigned
	cfg.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := cfg.Save(internal.WorkerConfigWritePath(root, a.WtBase, workerID)); err != nil {
		return fmt.Errorf("save worker config: %w", err)
	}

	record, err = internal.BindTaskToWorker(root, taskID, workerID, time.Now().UTC())
	if err != nil {
		return err
	}

	if err := a.RunWorkerOpen(workerID, provider, model, newWindow, provider != "", model != ""); err != nil {
		return err
	}
	cfg, _, err = internal.LoadWorkerConfigByID(root, a.WtBase, workerID)
	if err != nil {
		return err
	}
	cfg.TaskID = record.TaskID
	cfg.TaskPath = record.TaskPath
	cfg.Status = record.Status
	cfg.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := cfg.Save(internal.WorkerYAMLPath(filepath.Join(root, a.WtBase, workerID))); err != nil {
		return fmt.Errorf("save local worker config: %w", err)
	}

	if cfg.PaneID != "" {
		msg := fmt.Sprintf("[Task Assigned] Read worker.yaml first, then open %s/task.yaml and %s/context.md. Do not rely on this message for task details.", record.TaskPath, record.TaskPath)
		a.Session.PaneSend(cfg.PaneID, msg)
	}

	fmt.Printf("✓ Assigned task '%s' to worker '%s'\n", record.TaskID, workerID)
	return nil
}

func defaultWorkerProvider(provider string) string {
	if provider != "" {
		return provider
	}
	return "claude"
}
