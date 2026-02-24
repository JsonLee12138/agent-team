// cmd/assign.go
package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newAssignCmd() *cobra.Command {
	var model string
	var proposal string
	cmd := &cobra.Command{
		Use:   `assign <name> "<description>" [provider]`,
		Short: "Create an OpenSpec change and notify the role session",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := ""
			if len(args) > 2 {
				provider = args[2]
			}
			return GetApp(cmd).RunAssign(args[0], args[1], provider, model, proposal)
		},
	}
	cmd.Flags().StringVarP(&model, "model", "m", "", "AI model identifier")
	cmd.Flags().StringVarP(&proposal, "proposal", "p", "", "Path to proposal file (use - for stdin)")
	return cmd
}

func (a *App) RunAssign(name, desc, provider, model, proposalPath string) error {
	root := a.Git.Root()
	teamsDir := internal.TeamsDir(root, a.WtBase, name)
	configPath := internal.ConfigPath(root, a.WtBase, name)
	wtPath := internal.WtPath(root, a.WtBase, name)

	if _, err := os.Stat(teamsDir); os.IsNotExist(err) {
		return fmt.Errorf("role '%s' not found", name)
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

	// Create OpenSpec change
	ts := time.Now().Format("2006-01-02-15-04-05")
	slug := internal.Slugify(desc, 50)
	changeName := fmt.Sprintf("%s-%s", ts, slug)

	changePath, err := internal.CreateChange(wtPath, changeName, proposalContent)
	if err != nil {
		return err
	}
	fmt.Printf("✓ Change created: %s\n", changePath)

	// Ensure session is running
	cfg, err := internal.LoadRoleConfig(configPath)
	if err != nil {
		return err
	}

	if !a.Session.PaneAlive(cfg.PaneID) {
		fmt.Printf("Role '%s' is not running, opening session first...\n", name)
		if err := a.RunOpen(name, provider, model); err != nil {
			return err
		}
		cfg, err = internal.LoadRoleConfig(configPath)
		if err != nil {
			return err
		}
		fmt.Println("  Waiting for AI to initialize...")
		time.Sleep(3 * time.Second)
	}

	// Notify role
	changeRel := fmt.Sprintf("openspec/changes/%s/", changeName)
	msg := fmt.Sprintf("[New Change Assigned] %s\nChange: %s\nProposal ready. Run /opsx:continue to proceed.",
		desc, changeRel)
	a.Session.PaneSend(cfg.PaneID, msg)

	fmt.Printf("✓ Assigned to '%s': %s\n", name, desc)
	return nil
}
