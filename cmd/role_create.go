// cmd/role_create.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newRoleCreateCmd() *cobra.Command {
	var (
		description        string
		systemGoal         string
		inScope            []string
		outOfScope         []string
		skills             string
		recommendedSkills  string
		addSkills          string
		removeSkills       string
		manualSkills       string
		targetDir          string
		overwrite          string
		repoRoot           string
	)

	cmd := &cobra.Command{
		Use:   "create <role-name>",
		Short: "Create or update a role skill package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			roleName := args[0]

			validName, err := internal.ValidateRoleName(roleName)
			if err != nil {
				return err
			}

			finalSkills := internal.ResolveFinalSkills(
				internal.ParseCSVList(skills),
				internal.ParseCSVList(recommendedSkills),
				internal.ParseCSVList(addSkills),
				internal.ParseCSVList(removeSkills),
				internal.ParseCSVList(manualSkills),
			)

			root := repoRoot
			if root == "." || root == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("get working directory: %w", err)
				}
				root = cwd
			}

			config := internal.RoleConfig{
				RoleName:    validName,
				Description: strings.TrimSpace(description),
				SystemGoal:  strings.TrimSpace(systemGoal),
				InScope:     internal.CollectScope(inScope, strings.TrimSpace(description)),
				OutOfScope:  internal.CollectScope(outOfScope, "Tasks outside this role responsibilities"),
				Skills:      finalSkills,
			}

			confirmFn := func(targetDirPath string) (bool, error) {
				fmt.Fprintf(cmd.OutOrStdout(),
					"Role directory '%s' already exists. Overwrite managed files? [y/N]: ",
					targetDirPath)
				reader := bufio.NewReader(cmd.InOrStdin())
				answer, err := reader.ReadString('\n')
				if err != nil {
					return false, err
				}
				answer = strings.TrimSpace(strings.ToLower(answer))
				return answer == "y" || answer == "yes", nil
			}

			result, err := internal.CreateOrUpdateRole(
				root,
				config,
				overwrite,
				confirmFn,
				time.Now,
				targetDir,
			)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Generated role skill at %s\n", result.TargetDir)
			if result.BackupPath != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Backup created at %s\n", result.BackupPath)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Managed files: SKILL.md, references/role.yaml, system.md")
			return nil
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "Role description (required)")
	cmd.Flags().StringVar(&systemGoal, "system-goal", "", "Primary objective for the role system prompt (required)")
	cmd.Flags().StringArrayVar(&inScope, "in-scope", nil, "In-scope item (repeatable, supports comma-separated values)")
	cmd.Flags().StringArrayVar(&outOfScope, "out-of-scope", nil, "Out-of-scope item (repeatable, supports comma-separated values)")
	cmd.Flags().StringVar(&skills, "skills", "", "Final selected skills (comma-separated)")
	cmd.Flags().StringVar(&recommendedSkills, "recommended-skills", "", "Recommended skills from find-skills (comma-separated)")
	cmd.Flags().StringVar(&addSkills, "add-skills", "", "Skills to add (comma-separated)")
	cmd.Flags().StringVar(&removeSkills, "remove-skills", "", "Skills to remove from the candidate list")
	cmd.Flags().StringVar(&manualSkills, "manual-skills", "", "Manual fallback when recommendation is unavailable or empty")
	cmd.Flags().StringVar(&targetDir, "target-dir", "skills", "Target directory (skills | .agents/teams | custom path)")
	cmd.Flags().StringVar(&overwrite, "overwrite", "ask", "Overwrite mode: ask/yes/no")
	cmd.Flags().StringVar(&repoRoot, "repo-root", ".", "Repository root path")

	_ = cmd.MarkFlagRequired("description")
	_ = cmd.MarkFlagRequired("system-goal")

	return cmd
}
