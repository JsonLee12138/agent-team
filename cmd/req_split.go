// cmd/req_split.go
package cmd

import (
	"fmt"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newReqSplitCmd() *cobra.Command {
	var tasks []string

	cmd := &cobra.Command{
		Use:   "split <name>",
		Short: "Add sub-tasks to a requirement",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunReqSplit(args[0], tasks)
		},
	}

	cmd.Flags().StringArrayVar(&tasks, "task", nil, "Sub-task title (repeatable)")
	_ = cmd.MarkFlagRequired("task")

	return cmd
}

func (a *App) RunReqSplit(name string, tasks []string) error {
	root := a.Git.Root()

	req, err := internal.LoadRequirement(root, name)
	if err != nil {
		return fmt.Errorf("load requirement: %w", err)
	}

	// Find the next available ID
	nextID := 1
	for _, st := range req.SubTasks {
		if st.ID >= nextID {
			nextID = st.ID + 1
		}
	}

	for _, title := range tasks {
		req.SubTasks = append(req.SubTasks, internal.SubTask{
			ID:     nextID,
			Title:  title,
			Status: internal.SubTaskStatusPending,
		})
		fmt.Printf("  Added sub-task %d: %s\n", nextID, title)
		nextID++
	}

	if err := internal.SaveRequirement(root, req); err != nil {
		return fmt.Errorf("save requirement: %w", err)
	}

	if err := internal.UpdateIndexEntry(root, req); err != nil {
		return fmt.Errorf("update index: %w", err)
	}

	fmt.Printf("Split requirement '%s' into %d sub-tasks (total: %d)\n", name, len(tasks), len(req.SubTasks))
	return nil
}
