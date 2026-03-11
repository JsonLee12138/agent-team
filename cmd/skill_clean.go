// cmd/skill_clean.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newSkillCleanCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove cached skills from project-level cache directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := GetApp(cmd).Git.Root()
			wtBase := internal.FindWtBase(root)

			dir := filepath.Join(root, ".agents", ".cache", "skills")
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) == 0 {
				fmt.Println("No cached skills found.")
				return nil
			}

			// Check which cached skills are used by active worktrees
			usage := internal.FindCachedSkillUsage(root, wtBase)

			var free []string
			var inUse []string
			for _, e := range entries {
				if workers, ok := usage[e.Name()]; ok && len(workers) > 0 {
					inUse = append(inUse, e.Name())
				} else {
					free = append(free, e.Name())
				}
			}

			removed := 0

			// Clean unused skills without confirmation
			for _, name := range free {
				full := filepath.Join(dir, name)
				if err := os.RemoveAll(full); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to remove %s: %v\n", full, err)
				} else {
					removed++
					fmt.Printf("  Removed %s\n", name)
				}
			}

			// Handle in-use skills
			if len(inUse) > 0 {
				fmt.Printf("\nThe following cached skills are in use by active workers:\n")
				for _, name := range inUse {
					workers := usage[name]
					fmt.Printf("  %s  (used by: %s)\n", name, strings.Join(workers, ", "))
				}

				if force {
					fmt.Println("\n--force specified, removing in-use skills...")
				} else {
					fmt.Print("\nRemove these in-use skills? This will break symlinks in the workers above. [y/N]: ")
					reader := bufio.NewReader(os.Stdin)
					answer, _ := reader.ReadString('\n')
					answer = strings.TrimSpace(strings.ToLower(answer))
					if answer != "y" && answer != "yes" {
						if removed == 0 {
							fmt.Println("No skills removed.")
						} else {
							fmt.Printf("Cleaned %d cached skill(s), skipped %d in-use skill(s).\n", removed, len(inUse))
						}
						return nil
					}
				}

				for _, name := range inUse {
					full := filepath.Join(dir, name)
					if err := os.RemoveAll(full); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to remove %s: %v\n", full, err)
					} else {
						removed++
						fmt.Printf("  Removed %s\n", name)
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
	cmd.Flags().BoolVar(&force, "force", false, "Remove all cached skills without confirmation, including those in use")
	return cmd
}
