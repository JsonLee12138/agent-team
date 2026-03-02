// cmd/task_create.go
package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newTaskCreateCmd() *cobra.Command {
	var proposal string
	var design string
	var verifyCmd string
	var skipVerify bool

	cmd := &cobra.Command{
		Use:   `create <worker-id> "<description>"`,
		Short: "Create a new task change",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetApp(cmd).RunTaskCreate(args[0], args[1], proposal, design, verifyCmd, skipVerify)
		},
	}

	cmd.Flags().StringVarP(&proposal, "proposal", "p", "", "Path to proposal file (use - for stdin)")
	cmd.Flags().StringVarP(&design, "design", "d", "", "Path to design file")
	cmd.Flags().StringVar(&verifyCmd, "verify-cmd", "", "Command to verify the change")
	cmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "Skip verification")

	return cmd
}

func (a *App) RunTaskCreate(workerID, description, proposalPath, designPath, verifyCmd string, skipVerify bool) error {
	root := a.Git.Root()
	wtPath := internal.WtPath(root, a.WtBase, workerID)

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("worker '%s' not found at %s", workerID, wtPath)
	}

	// Read proposal content
	var proposalContent string
	if proposalPath != "" {
		var data []byte
		var err error
		if proposalPath == "-" {
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(proposalPath)
		}
		if err != nil {
			return fmt.Errorf("read proposal: %w", err)
		}
		proposalContent = string(data)
	}

	// Read design content
	var designContent string
	if designPath != "" {
		data, err := os.ReadFile(designPath)
		if err != nil {
			return fmt.Errorf("read design: %w", err)
		}
		designContent = string(data)
	}

	// Generate change name from timestamp + description
	ts := time.Now().Format("2006-01-02-15-04-05")
	slug := internal.Slugify(description, 50)
	changeName := fmt.Sprintf("%s-%s", ts, slug)

	// Create the change
	changeDir, err := internal.CreateTaskChange(wtPath, changeName, description, proposalContent, designContent)
	if err != nil {
		return err
	}
	fmt.Printf("✓ Change created: %s\n", changeDir)

	// Update verify config if specified
	if verifyCmd != "" || skipVerify {
		change, err := internal.LoadChange(wtPath, changeName)
		if err != nil {
			return fmt.Errorf("load change: %w", err)
		}

		if skipVerify {
			change.Verify.Skip = true
		}
		if verifyCmd != "" {
			change.Verify.Command = verifyCmd
		}

		if err := internal.SaveChange(wtPath, change); err != nil {
			return fmt.Errorf("save change: %w", err)
		}
	}

	fmt.Printf("✓ Created change: %s\n", changeName)
	fmt.Printf("  Description: %s\n", description)
	fmt.Printf("  Status: draft\n")
	return nil
}
