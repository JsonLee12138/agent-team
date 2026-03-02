// cmd/hook.go
package cmd

import "github.com/spf13/cobra"

func newHookCmd() *cobra.Command {
	var providerFlag string

	cmd := &cobra.Command{
		Use:    "hook",
		Short:  "Hook event handlers (called by provider plugins)",
		Hidden: true,
	}

	cmd.PersistentFlags().StringVar(&providerFlag, "provider", "auto", "Provider override (auto|claude|gemini|opencode)")

	cmd.AddCommand(
		newHookSessionStartCmd(),
		newHookStopCmd(),
		newHookPreToolUseCmd(),
		newHookPostToolUseCmd(),
		newHookTaskCompletedCmd(),
		newHookSubagentStartCmd(),
		newHookTeammateIdleCmd(),
	)

	return cmd
}
