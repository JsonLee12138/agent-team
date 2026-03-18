package cmd

import (
	"fmt"
	"os"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	var skipDetect bool

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Set up agent-team global environment",
		Long: `Detects installed providers, installs bundled plugin roles to the global scope,
and ensures the global roles directory exists.

This command replaces the global-installation part of the old 'init' command.
Use 'agent-team init' for project-level initialization (rules, provider files).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup(skipDetect)
		},
	}
	cmd.Flags().BoolVar(&skipDetect, "skip-detect", false, "Skip provider detection")
	return cmd
}

func runSetup(skipDetect bool) error {
	fmt.Println("agent-team setup")
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

	// Step 2: Ensure global roles directory
	globalDir, err := internal.EnsureGlobalRolesDir()
	if err != nil {
		return err
	}

	// Step 3: Scan and install plugin bundled roles to global
	candidates := internal.ScanPluginRoles()
	if len(candidates) > 0 {
		fmt.Printf("Bundled roles found: %d\n", len(candidates))
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
		fmt.Println("No bundled roles found in plugin.")
	}

	// Step 4: Summary
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Initialize a project: agent-team init")
	fmt.Println("  2. Find roles:           agent-team role-repo find <query>")
	fmt.Println("  3. List roles:           agent-team role-repo list")

	return nil
}
