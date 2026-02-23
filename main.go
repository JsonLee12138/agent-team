// main.go
package main

import (
	"os"

	"github.com/leeforge/agent-team/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	cmd.RegisterCommands(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
