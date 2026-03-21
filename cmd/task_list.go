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
	cmd.Flags().BoolVar(&archived, "archived", false, "Include archived tasks")
	return cmd
}

func (a *App) RunTaskList(includeArchived bool) error {
	tasks, err := internal.ListTasks(a.Git.Root(), !includeArchived)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	fmt.Printf("%-32s %-12s %-20s %s\n", "Task", "Status", "Role", "Worker")
	fmt.Printf("%-32s %-12s %-20s %s\n", "────────────────────────────────", "────────────", "────────────────────", "────────────────────────")
	for _, task := range tasks {
		worker := task.WorkerID
		if worker == "" {
			worker = "-"
		}
		fmt.Printf("%-32s %-12s %-20s %s\n", task.TaskID, task.Status, task.Role, worker)
	}
	return nil
}
