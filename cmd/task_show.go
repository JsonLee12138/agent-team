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
	record, location, err := internal.LoadTaskRecord(root, taskID)
	if err != nil {
		return err
	}
	contextPath := taskContextPathByLocation(root, taskID, location)
	verificationPath := taskVerificationPathByLocation(root, taskID, location)
	contextData, err := os.ReadFile(contextPath)
	if err != nil {
		return fmt.Errorf("read context.md: %w", err)
	}
	verificationData, err := os.ReadFile(verificationPath)
	if err != nil {
		return fmt.Errorf("read verification.md: %w", err)
	}
	verificationResult := internal.ParseVerificationResult(string(verificationData))

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
	if record.VerifyingAt != "" {
		fmt.Printf("Verifying At: %s\n", record.VerifyingAt)
	}
	if record.ArchivedAt != "" {
		fmt.Printf("Archived At: %s\n", record.ArchivedAt)
	}
	if record.DeprecatedAt != "" {
		fmt.Printf("Deprecated At: %s\n", record.DeprecatedAt)
	}
	if record.MergedSHA != "" {
		fmt.Printf("Merged SHA: %s\n", record.MergedSHA)
	}
	fmt.Printf("Verification Exists: yes\n")
	fmt.Printf("Verification Result: %s\n", verificationResult)
	fmt.Printf("Archive Ready (default): %s\n", yesNo(canArchive(verificationResult, false)))
	fmt.Printf("Archive Ready (strict): %s\n", yesNo(canArchive(verificationResult, true)))
	fmt.Printf("\n%s\n", strings.TrimSpace(string(contextData)))
	fmt.Printf("\n\n%s\n", strings.TrimSpace(string(verificationData)))
	return nil
}

func taskContextPathByLocation(root, taskID string, location internal.TaskRecordLocation) string {
	switch location {
	case internal.TaskRecordLocationArchived:
		return internal.TaskArchiveContextPath(root, taskID)
	case internal.TaskRecordLocationDeprecated:
		return internal.TaskDeprecatedContextPath(root, taskID)
	default:
		return internal.TaskContextPath(root, taskID)
	}
}

func taskVerificationPathByLocation(root, taskID string, location internal.TaskRecordLocation) string {
	switch location {
	case internal.TaskRecordLocationArchived:
		return internal.TaskArchiveVerificationPath(root, taskID)
	case internal.TaskRecordLocationDeprecated:
		return internal.TaskDeprecatedVerificationPath(root, taskID)
	default:
		return internal.TaskVerificationPath(root, taskID)
	}
}

func canArchive(result internal.VerificationResult, strict bool) bool {
	return internal.ValidateArchiveReadiness(result, strict) == nil
}

func yesNo(ok bool) string {
	if ok {
		return "yes"
	}
	return "no"
}
