// cmd/skill_check.go
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func newSkillCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check installed skills for available updates",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := GetApp(cmd).Git.Root()
			fmt.Println("Checking skills in project root:", root)
			c := exec.Command("npx", "skills", "check")
			c.Dir = root
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}
}
