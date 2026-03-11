// cmd/skill_clean.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newSkillCleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Remove all cached skills from project-level cache directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := GetApp(cmd).Git.Root()

			// Clean skill cache from .agents/.cache/skills/
			dir := filepath.Join(root, ".agents", ".cache", "skills")
			removed := 0
			entries, err := os.ReadDir(dir)
			if err == nil {
				for _, e := range entries {
					full := filepath.Join(dir, e.Name())
					if err := os.RemoveAll(full); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to remove %s: %v\n", full, err)
					} else {
						removed++
						fmt.Printf("  Removed %s\n", full)
					}
				}
			}

			if removed == 0 {
				fmt.Println("No cached skills found.")
			} else {
				fmt.Printf("Cleaned %d cached skill(s).\n", removed)
			}
			return nil
		},
	}
}
