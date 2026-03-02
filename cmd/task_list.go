// cmd/task_list.go
package cmd

import (
	"fmt"
	"os"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newTaskListCmd() *cobra.Command {
	var status string
	var worker string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all task changes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunTaskList(worker, status)
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status (draft|assigned|implementing|verifying|done|archived)")
	cmd.Flags().StringVarP(&worker, "worker", "w", "", "Filter by worker ID")

	return cmd
}

func (a *App) RunTaskList(workerID, statusFilter string) error {
	root := a.Git.Root()

	// Determine which worker(s) to list
	var workers []internal.WorkerInfo
	if workerID != "" {
		wtPath := internal.WtPath(root, a.WtBase, workerID)
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			return fmt.Errorf("worker '%s' not found", workerID)
		}
		cfg, err := internal.LoadWorkerConfig(internal.WorkerYAMLPath(wtPath))
		if err != nil {
			return fmt.Errorf("load worker config: %w", err)
		}
		workers = append(workers, internal.WorkerInfo{
			WorkerID: workerID,
			Config:   cfg,
		})
	} else {
		workers = internal.ListWorkers(root, a.WtBase)
	}

	if len(workers) == 0 {
		fmt.Println("No workers found.")
		return nil
	}

	fmt.Printf("%-36s %-16s %-16s %-12s %s\n", "Change", "Status", "Worker", "Tasks", "CreatedAt")
	fmt.Printf("%-36s %-16s %-16s %-12s %s\n",
		"────────────────────────────────────",
		"────────────────",
		"────────────────",
		"────────────",
		"──────────────────────")

	for _, worker := range workers {
		if worker.Config == nil {
			continue
		}
		wtPath := internal.WtPath(root, a.WtBase, worker.WorkerID)

		changes, err := internal.ListChanges(wtPath)
		if err != nil {
			continue
		}

		for _, change := range changes {
			// Filter by status if specified
			if statusFilter != "" && string(change.Status) != statusFilter {
				continue
			}

			taskCount := len(change.Tasks)
			tasksSummary := fmt.Sprintf("%d", taskCount)

			fmt.Printf("%-36s %-16s %-16s %-12s %s\n",
				change.Name,
				change.Status,
				worker.WorkerID,
				tasksSummary,
				change.CreatedAt)
		}
	}

	return nil
}
