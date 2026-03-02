// cmd/migrate.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrate agents/ directory to .agents/ (hidden)",
		Long:  "Moves the agents/ directory to .agents/ for Claude Code Plugin compatibility.",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := GetApp(cmd)
			return runMigrate(app.Git.Root())
		},
	}
}

func runMigrate(root string) error {
	oldPath := root + "/agents"
	newPath := root + "/.agents"

	// 检测 agents/ 是否存在
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		fmt.Println("Nothing to migrate: agents/ directory not found.")
		return nil
	}

	// 检测 .agents/ 是否已存在，防止数据丢失
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf(".agents/ already exists — remove it manually before migrating to avoid data loss")
	}

	// 原子移动
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("migrate failed: %w", err)
	}

	fmt.Printf("✓ Migrated agents/ → .agents/\n")
	fmt.Println("  You can now use .agents/ for all agent-team operations.")
	return nil
}
