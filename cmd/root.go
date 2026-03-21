// cmd/root.go
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

var Version = "dev"

type App struct {
	Git     *internal.GitClient
	Session internal.SessionBackend
	WtBase  string
}

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "agent-team",
		Short:   "AI team role manager",
		Long:    "Manages AI team roles using git worktrees and terminal multiplexer tabs.",
		Version: Version,
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Skip app bootstrap only for root-level utility commands.
		switch cmd.CommandPath() {
		case "agent-team help", "agent-team version", "agent-team completion", "agent-team _inject-role-prompt", "agent-team _record-main-pane", "agent-team init", "agent-team setup":
			return nil
		}
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}

		gc, err := internal.NewGitClient(cwd)
		if err != nil {
			return fmt.Errorf("not in a git repository")
		}

		branch, err := gc.CurrentBranch()
		if err != nil {
			branch = "main"
		}
		requiresInitialization := !strings.HasPrefix(branch, "team/")

		// Check if .agent-team/rules/ exists (initialization check)
		if requiresInitialization && !internal.HasRulesDir(gc.Root()) {
			// Check if running in non-interactive mode
			nonInteractive := os.Getenv("AGENT_TEAM_NONINTERACTIVE") == "1"

			if nonInteractive {
				return fmt.Errorf(".agent-team/rules/ not found. Run 'agent-team init' first (or set AGENT_TEAM_NONINTERACTIVE=0 for interactive mode)")
			}

			// Interactive mode: prompt user to run init
			fmt.Println("⚠️  .agent-team/rules/ not found. This project has not been initialized for agent-team.")
			fmt.Print("Run agent-team init now? [Y/n] ")

			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("read input: %w", err)
			}

			response := strings.TrimSpace(strings.ToLower(input))
			if response == "" || response == "y" || response == "yes" {
				fmt.Println()
				if err := runInit(); err != nil {
					return fmt.Errorf("initialization failed: %w", err)
				}
				fmt.Println()
			} else {
				return fmt.Errorf("initialization cancelled by user")
			}
		}

		app := &App{
			Git:     gc,
			Session: internal.NewSessionBackend(),
			WtBase:  internal.FindWtBase(gc.Root()),
		}
		cmd.SetContext(WithApp(cmd.Context(), app))
		return nil
	}

	return rootCmd
}

// Context key for App
type appKey struct{}

func WithApp(ctx context.Context, app *App) context.Context {
	return context.WithValue(ctx, appKey{}, app)
}

func GetApp(cmd *cobra.Command) *App {
	return cmd.Context().Value(appKey{}).(*App)
}

func RegisterCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newWorkerCmd())
	rootCmd.AddCommand(newRoleCmd())
	rootCmd.AddCommand(newRoleRepoCmd())
	rootCmd.AddCommand(newReplyCmd())
	rootCmd.AddCommand(newReplyMainCmd())
	rootCmd.AddCommand(newCompactCmd())
	rootCmd.AddCommand(newContextCleanupCmd())
	rootCmd.AddCommand(newMigrateCmd())
	rootCmd.AddCommand(newInjectRolePromptCmd())
	rootCmd.AddCommand(newRecordMainPaneCmd())
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newSetupCmd())
	rootCmd.AddCommand(newRulesCmd())
	rootCmd.AddCommand(newCatalogCmd())
	rootCmd.AddCommand(newSkillCmd())
	rootCmd.AddCommand(newWorkflowCmd())
	rootCmd.AddCommand(newTaskCmd())
}
