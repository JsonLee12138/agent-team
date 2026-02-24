// main.go
package main

import (
	"os"

	"github.com/JsonLee12138/agent-team/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	cmd.RegisterCommands(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
