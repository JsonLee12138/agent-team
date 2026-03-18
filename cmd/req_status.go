// cmd/req_status.go
package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newReqStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [name]",
		Short: "Show requirement status",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return GetApp(cmd).RunReqStatusAll(cmd)
			}
			return GetApp(cmd).RunReqStatusOne(cmd, args[0])
		},
	}

	return cmd
}

func (a *App) RunReqStatusAll(cmd *cobra.Command) error {
	root := a.Git.Root()

	reqs, err := internal.ListRequirements(root)
	if err != nil {
		return fmt.Errorf("list requirements: %w", err)
	}

	if len(reqs) == 0 {
		fmt.Println("No requirements found.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tPROGRESS\tSUB-TASKS")

	for _, req := range reqs {
		total := len(req.SubTasks)
		done := 0
		for _, st := range req.SubTasks {
			if st.Status == internal.SubTaskStatusDone || st.Status == internal.SubTaskStatusSkipped {
				done++
			}
		}

		progress := "0%"
		if total > 0 {
			progress = fmt.Sprintf("%d%%", done*100/total)
		}

		fmt.Fprintf(w, "%s\t%s\t%s (%d/%d)\t%d\n",
			req.Name, req.Status, progress, done, total, total)
	}

	return w.Flush()
}

func (a *App) RunReqStatusOne(cmd *cobra.Command, name string) error {
	root := a.Git.Root()

	req, err := internal.LoadRequirement(root, name)
	if err != nil {
		return fmt.Errorf("load requirement: %w", err)
	}

	total := len(req.SubTasks)
	done := 0
	for _, st := range req.SubTasks {
		if st.Status == internal.SubTaskStatusDone || st.Status == internal.SubTaskStatusSkipped {
			done++
		}
	}

	progress := "0%"
	if total > 0 {
		progress = fmt.Sprintf("%d%%", done*100/total)
	}

	fmt.Printf("Requirement: %s\n", req.Name)
	fmt.Printf("Description: %s\n", req.Description)
	fmt.Printf("Status:      %s\n", req.Status)
	fmt.Printf("Progress:    %s (%d/%d)\n", progress, done, total)
	fmt.Println()

	if total == 0 {
		fmt.Println("No sub-tasks.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tASSIGNED_TO\tCHANGE")

	for _, st := range req.SubTasks {
		assignedTo := st.AssignedTo
		if assignedTo == "" {
			assignedTo = "-"
		}
		changeName := st.ChangeName
		if changeName == "" {
			changeName = "-"
		}
		// Truncate long change names for display
		if len(changeName) > 40 {
			changeName = changeName[:37] + "..."
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			st.ID, st.Title, st.Status, assignedTo, changeName)
	}

	return w.Flush()
}
