// cmd/root.go
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/leeforge/agent-team/internal"
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
		// Skip init for help/version/completion
		if cmd.Name() == "help" || cmd.Name() == "version" || cmd.Name() == "completion" {
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
	// Commands will be added in subsequent tasks
}
