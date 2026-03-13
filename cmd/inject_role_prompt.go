// cmd/inject_role_prompt.go
package cmd

import (
	"fmt"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

// newInjectRolePromptCmd 返回隐藏子命令，供 worker open 流程调用。
func newInjectRolePromptCmd() *cobra.Command {
	var worktree, workerID, role, root string

	cmd := &cobra.Command{
		Use:    "_inject-role-prompt",
		Hidden: true,
		Short:  "Inject role prompt into worktree CLAUDE.md (internal)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if worktree == "" || workerID == "" || role == "" || root == "" {
				return fmt.Errorf("--worktree, --worker-id, --role, and --root are all required")
			}
			// Try to read RolePath from worker.yaml for global role support
			configPath := internal.WorkerYAMLPath(worktree)
			if cfg, err := internal.LoadWorkerConfig(configPath); err == nil && cfg.RolePath != "" {
				return internal.InjectRolePromptWithPath(worktree, workerID, role, cfg.RolePath, root)
			}
			return internal.InjectRolePrompt(worktree, workerID, role, root)
		},
	}

	cmd.Flags().StringVar(&worktree, "worktree", "", "Path to the git worktree directory")
	cmd.Flags().StringVar(&workerID, "worker-id", "", "Worker ID (e.g. frontend-dev-001)")
	cmd.Flags().StringVar(&role, "role", "", "Role name (e.g. frontend-dev)")
	cmd.Flags().StringVar(&root, "root", "", "Main project root directory")

	return cmd
}
