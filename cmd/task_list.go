package cmd

import (
	"fmt"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newTaskListCmd() *cobra.Command {
	var archived bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunTaskList(archived)
		},
	}
	cmd.Flags().BoolVar(&archived, "archived", false, "Include archived and deprecated tasks")
	return cmd
}

func (a *App) RunTaskList(includeArchived bool) error {
	root := a.Git.Root()
	tasks, err := internal.ListTasks(root, !includeArchived)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	fmt.Printf("%-32s %-12s %-20s %-12s %-14s %s\n", "Task", "Status", "Role", "Verification", "Archive Ready", "Worker")
	fmt.Printf("%-32s %-12s %-20s %-12s %-14s %s\n", "────────────────────────────────", "────────────", "────────────────────", "────────────", "──────────────", "────────────────────────")
	for _, task := range tasks {
		worker := task.WorkerID
		if worker == "" {
			worker = "-"
		}
		verification := internal.VerificationResultMissing
		if result, err := internal.ReadTaskVerificationResult(root, task.TaskID, taskLocationForStatus(task.Status)); err == nil {
			verification = result
		}
		fmt.Printf("%-32s %-12s %-20s %-12s %-14s %s\n", task.TaskID, task.Status, task.Role, verification, internal.ArchiveReadyLabel(verification), worker)
	}
	return nil
}

func taskLocationForStatus(status internal.TaskStatus) internal.TaskRecordLocation {
	switch status {
	case internal.TaskStatusArchived:
		return internal.TaskRecordLocationArchived
	case internal.TaskStatusDeprecated:
		return internal.TaskRecordLocationDeprecated
	default:
		return internal.TaskRecordLocationActive
	}
}
