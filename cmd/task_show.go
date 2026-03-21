package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newTaskShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <task-id>",
		Short: "Show task package details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunTaskShow(args[0])
		},
	}
	return cmd
}

func (a *App) RunTaskShow(taskID string) error {
	root := a.Git.Root()
	record, archived, err := internal.LoadTaskRecord(root, taskID)
	if err != nil {
		return err
	}
	contextPath := internal.TaskContextPath(root, taskID)
	verificationPath := internal.TaskVerificationPath(root, taskID)
	if archived {
		contextPath = internal.TaskArchiveContextPath(root, taskID)
		verificationPath = internal.TaskArchiveVerificationPath(root, taskID)
	}
	contextData, err := os.ReadFile(contextPath)
	if err != nil {
		return fmt.Errorf("read context.md: %w", err)
	}
	verificationData, err := os.ReadFile(verificationPath)
	if err != nil {
		return fmt.Errorf("read verification.md: %w", err)
	}

	fmt.Printf("Task: %s\n", record.TaskID)
	fmt.Printf("Title: %s\n", record.Title)
	fmt.Printf("Role: %s\n", record.Role)
	fmt.Printf("Status: %s\n", record.Status)
	fmt.Printf("Task Path: %s\n", record.TaskPath)
	if record.WorkerID != "" {
		fmt.Printf("Worker: %s\n", record.WorkerID)
	}
	fmt.Printf("Created At: %s\n", record.CreatedAt)
	if record.AssignedAt != "" {
		fmt.Printf("Assigned At: %s\n", record.AssignedAt)
	}
	if record.DoneAt != "" {
		fmt.Printf("Done At: %s\n", record.DoneAt)
	}
	if record.ArchivedAt != "" {
		fmt.Printf("Archived At: %s\n", record.ArchivedAt)
	}
	if record.MergedSHA != "" {
		fmt.Printf("Merged SHA: %s\n", record.MergedSHA)
	}
	fmt.Printf("\n%s\n", strings.TrimSpace(string(contextData)))
	fmt.Printf("\n\n%s\n", strings.TrimSpace(string(verificationData)))
	return nil
}
