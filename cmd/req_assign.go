// cmd/req_assign.go
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newReqAssignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign <name> <task-id> <worker-id>",
		Short: "Assign a sub-task to a worker",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid task-id: %s", args[1])
			}
			return GetApp(cmd).RunReqAssign(args[0], taskID, args[2])
		},
	}

	return cmd
}

func (a *App) RunReqAssign(name string, taskID int, workerID string) error {
	root := a.Git.Root()

	// Verify worker exists
	wtPath := internal.WtPath(root, a.WtBase, workerID)
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("worker '%s' not found at %s", workerID, wtPath)
	}

	req, err := internal.LoadRequirement(root, name)
	if err != nil {
		return fmt.Errorf("load requirement: %w", err)
	}

	// Find the sub-task title for the change description
	var taskTitle string
	for _, st := range req.SubTasks {
		if st.ID == taskID {
			taskTitle = st.Title
			break
		}
	}
	if taskTitle == "" {
		return fmt.Errorf("sub-task %d not found in requirement '%s'", taskID, name)
	}

	// Create a Change for this assignment
	ts := time.Now().Format("2006-01-02-15-04-05")
	slug := internal.Slugify(taskTitle, 50)
	changeName := fmt.Sprintf("%s-%s", ts, slug)
	description := fmt.Sprintf("req/%s#%d: %s", name, taskID, taskTitle)

	_, err = internal.CreateTaskChange(wtPath, changeName, description, "", "")
	if err != nil {
		return fmt.Errorf("create change: %w", err)
	}

	// Assign the sub-task
	if err := internal.AssignSubTask(req, taskID, workerID, changeName); err != nil {
		return fmt.Errorf("assign sub-task: %w", err)
	}

	if err := internal.SaveRequirement(root, req); err != nil {
		return fmt.Errorf("save requirement: %w", err)
	}

	if err := internal.UpdateIndexEntry(root, req); err != nil {
		return fmt.Errorf("update index: %w", err)
	}

	fmt.Printf("Assigned sub-task %d to worker '%s'\n", taskID, workerID)
	fmt.Printf("  Change: %s\n", changeName)
	fmt.Printf("  Requirement status: %s\n", req.Status)
	return nil
}
