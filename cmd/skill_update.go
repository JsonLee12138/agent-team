// cmd/skill_update.go
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func newSkillUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update all cached skills to latest versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := GetApp(cmd).Git.Root()
			fmt.Println("Updating skills in project root:", root)
			c := exec.Command("npx", "skills", "update")
			c.Dir = root
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}
}
