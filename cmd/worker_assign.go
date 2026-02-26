// cmd/worker_assign.go
package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newWorkerAssignCmd() *cobra.Command {
	var model string
	var proposal string
	var design string
	var newWindow bool
	cmd := &cobra.Command{
		Use:   `assign <worker-id> "<description>" [provider]`,
		Short: "Create an OpenSpec change and notify the worker session",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := ""
			if len(args) > 2 {
				provider = args[2]
			}
			return GetApp(cmd).RunWorkerAssign(args[0], args[1], provider, model, proposal, design, newWindow)
		},
	}
	cmd.Flags().StringVarP(&model, "model", "m", "", "AI model identifier")
	cmd.Flags().StringVarP(&proposal, "proposal", "p", "", "Path to proposal file (use - for stdin)")
	cmd.Flags().StringVarP(&design, "design", "d", "", "Path to design file (brainstorming output)")
	cmd.Flags().BoolVarP(&newWindow, "new-window", "w", false, "Open in a new window instead of a tab")
	return cmd
}

func (a *App) RunWorkerAssign(workerID, desc, provider, model, proposalPath, designPath string, newWindow bool) error {
	root := a.Git.Root()
	configPath := internal.WorkerConfigPath(root, workerID)
	wtPath := internal.WtPath(root, a.WtBase, workerID)

	cfg, err := internal.LoadWorkerConfig(configPath)
	if err != nil {
		return fmt.Errorf("worker '%s' not found: %w", workerID, err)
	}

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		return fmt.Errorf("worker worktree '%s' not found at %s", workerID, wtPath)
	}

	// Read proposal content
	var proposalContent string
	if proposalPath != "" {
		var data []byte
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

	// Create OpenSpec change
	ts := time.Now().Format("2006-01-02-15-04-05")
	slug := internal.Slugify(desc, 50)
	changeName := fmt.Sprintf("%s-%s", ts, slug)

	changePath, err := internal.CreateChange(wtPath, changeName, proposalContent, designContent)
	if err != nil {
		return err
	}
	fmt.Printf("✓ Change created: %s\n", changePath)

	// Ensure session is running
	if !a.Session.PaneAlive(cfg.PaneID) {
		fmt.Printf("Worker '%s' is not running, opening session first...\n", workerID)
		if err := a.RunWorkerOpen(workerID, provider, model, newWindow); err != nil {
			return err
		}
		cfg, err = internal.LoadWorkerConfig(configPath)
		if err != nil {
			return err
		}
		fmt.Println("  Waiting for AI to initialize...")
		time.Sleep(3 * time.Second)
	}

	// Notify worker
	changeRel := fmt.Sprintf("openspec/changes/%s/", changeName)
	msg := fmt.Sprintf("[New Change Assigned] %s\nChange: %s\nProposal ready. Run /opsx:continue to proceed.",
		desc, changeRel)
	a.Session.PaneSend(cfg.PaneID, msg)

	fmt.Printf("✓ Assigned to worker '%s': %s\n", workerID, desc)
	return nil
}
