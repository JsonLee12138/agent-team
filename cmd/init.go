package cmd

import (
	"fmt"
	"os"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var globalOnly bool
	var skipDetect bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize agent-team for this project",
		Long:  "Sets up project structure, detects installed providers, and installs bundled roles to the global scope.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(globalOnly, skipDetect)
		},
	}
	cmd.Flags().BoolVar(&globalOnly, "global-only", false, "Only install global roles, skip project setup")
	cmd.Flags().BoolVar(&skipDetect, "skip-detect", false, "Skip provider detection")
	return cmd
}

func runInit(globalOnly bool, skipDetect bool) error {
	fmt.Println("agent-team init")
	fmt.Println()

	// Step 1: Detect providers
	if !skipDetect {
		providers := internal.DetectInstalledProviders()
		if len(providers) > 0 {
			fmt.Printf("Detected providers: %s\n", internal.FormatProviderList(providers))
		} else {
			fmt.Println("No agent providers detected (claude, gemini, opencode, codex).")
			fmt.Println("Install at least one provider to use agent-team workers.")
		}
		fmt.Println()
	}

	// Step 2: Initialize project structure
	if !globalOnly {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		if err := internal.InitProject(cwd); err != nil {
			return err
		}
		fmt.Println("Project initialized: .agents/teams/ created")
	}

	// Step 3: Scan and install plugin bundled roles to global
	globalDir, err := internal.EnsureGlobalRolesDir()
	if err != nil {
		return err
	}

	candidates := internal.ScanPluginRoles()
	if len(candidates) > 0 {
		fmt.Printf("\nBundled roles found: %d\n", len(candidates))
		installed := 0
		updated := 0
		skipped := 0
		for _, c := range candidates {
			action, err := internal.InstallPluginRoleToGlobal(c, globalDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  - error installing %s: %v\n", c.Name, err)
				continue
			}
			switch action {
			case internal.InstallActionInstalled:
				fmt.Printf("  + installed %s\n", c.Name)
				installed++
			case internal.InstallActionUpdated:
				fmt.Printf("  ~ updated %s\n", c.Name)
				updated++
			case internal.InstallActionSkipped:
				skipped++
			}
		}
		if installed > 0 || updated > 0 {
			fmt.Printf("  Roles: %d installed, %d updated, %d up-to-date\n", installed, updated, skipped)
		} else {
			fmt.Printf("  All %d bundled role(s) up-to-date\n", skipped)
		}
	} else {
		fmt.Println("\nNo bundled roles found in plugin.")
	}

	// Step 4: Summary
	fmt.Println("\nNext steps:")
	if !globalOnly {
		fmt.Println("  1. Create roles:   agent-team role-repo find <query>")
		fmt.Println("  2. Create workers: agent-team worker create <role>")
	} else {
		fmt.Println("  1. Find roles:     agent-team role-repo find <query>")
	}
	fmt.Println("  3. List roles:     agent-team role-repo list")

	return nil
}
