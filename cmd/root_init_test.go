package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

// TestPersistentPreRunE_UninitializedProject tests the initialization check
// in PersistentPreRunE when .agents/rules/ does not exist.
func TestPersistentPreRunE_UninitializedProject(t *testing.T) {
	t.Run("non-interactive mode returns error", func(t *testing.T) {
		// Create a temp directory without .agents/rules/
		tmpDir := t.TempDir()

		// Initialize git repo (required for NewGitClient)
		runGitCommand(t, tmpDir, "init")

		// Save and restore cwd
		origCwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd: %v", err)
		}
		defer os.Chdir(origCwd) // nolint:errcheck
		os.Chdir(tmpDir)        // nolint:errcheck

		// Set non-interactive mode AFTER changing directory
		t.Setenv("AGENT_TEAM_NONINTERACTIVE", "1")

		// Verify env var is set
		if os.Getenv("AGENT_TEAM_NONINTERACTIVE") != "1" {
			t.Fatalf("AGENT_TEAM_NONINTERACTIVE is '%s', expected '1'", os.Getenv("AGENT_TEAM_NONINTERACTIVE"))
		}

		// Verify .agents/rules/ does not exist
		if _, err := os.Stat(filepath.Join(tmpDir, ".agents", "rules")); err == nil {
			t.Fatal(".agents/rules/ should not exist")
		}

		// Debug: check what HasRulesDir returns
		t.Logf("HasRulesDir(tmpDir) = %v", internal.HasRulesDir(tmpDir))

		// Create root cmd and worker subcommand
		rootCmd := NewRootCmd()
		workerCmd := &cobra.Command{
			Use: "worker",
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}
		rootCmd.AddCommand(workerCmd)

		// Manually invoke PersistentPreRunE
		if rootCmd.PersistentPreRunE == nil {
			t.Fatal("PersistentPreRunE is not set")
		}

		// Execute PersistentPreRunE directly
		err = rootCmd.PersistentPreRunE(workerCmd, []string{})

		// The error should be about initialization
		if err == nil {
			t.Fatal("Expected error when .agents/rules/ not found in non-interactive mode")
		}
		t.Logf("Got error: %v", err)
		if !contains(err.Error(), ".agents/rules/ not found") {
			t.Errorf("Expected initialization error, got: %v", err)
		}
	})

	t.Run("interactive mode would prompt (skipped)", func(t *testing.T) {
		// Interactive mode requires stdin input, which is hard to test
		// This test documents the expected behavior
		t.Skip("Interactive mode test requires mocking stdin")
	})
}

// TestPersistentPreRunE_InitializedProject tests that Initialized project
// passes PersistentPreRunE without error.
func TestPersistentPreRunE_InitializedProject(t *testing.T) {
	t.Run("passes when .agents/rules/ exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize git repo
		runGitCommand(t, tmpDir, "init")

		// Create .agents/rules/ directory
		rulesDir := filepath.Join(tmpDir, ".agents", "rules")
		if err := os.MkdirAll(rulesDir, 0755); err != nil {
			t.Fatalf("create rules dir: %v", err)
		}

		// Save and restore cwd
		origCwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd: %v", err)
		}
		defer os.Chdir(origCwd) // nolint:errcheck
		os.Chdir(tmpDir)        // nolint:errcheck

		// Create root cmd and try to execute a command
		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"worker", "list"})

		// Should not fail due to initialization check
		// (may fail for other reasons like no worktrees, but that's expected)
		err = rootCmd.Execute()
		// We expect this to at least pass the initialization check
		// It might fail later in the command execution, which is fine
		if err != nil {
			// Check if error is NOT about initialization
			if contains(err.Error(), ".agents/rules/") {
				t.Errorf("Should pass initialization check, got: %v", err)
			}
		}
	})
}

// TestHasRulesDir_Integration tests HasRulesDir helper in various scenarios.
func TestHasRulesDir_Integration(t *testing.T) {
	t.Run("returns true when .agents/rules/ exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		rulesDir := filepath.Join(tmpDir, ".agents", "rules")
		if err := os.MkdirAll(rulesDir, 0755); err != nil {
			t.Fatalf("create rules dir: %v", err)
		}

		if !internal.HasRulesDir(tmpDir) {
			t.Error("HasRulesDir should return true when .agents/rules/ exists")
		}
	})

	t.Run("returns false when .agents/ exists but rules/ does not", func(t *testing.T) {
		tmpDir := t.TempDir()
		agentsDir := filepath.Join(tmpDir, ".agents")
		if err := os.MkdirAll(agentsDir, 0755); err != nil {
			t.Fatalf("create agents dir: %v", err)
		}

		if internal.HasRulesDir(tmpDir) {
			t.Error("HasRulesDir should return false when rules/ does not exist")
		}
	})

	t.Run("returns false when .agents/ does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		if internal.HasRulesDir(tmpDir) {
			t.Error("HasRulesDir should return false when .agents/ does not exist")
		}
	})

	t.Run("returns false when rules is a file not a directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		agentsDir := filepath.Join(tmpDir, ".agents")
		if err := os.MkdirAll(agentsDir, 0755); err != nil {
			t.Fatalf("create agents dir: %v", err)
		}

		rulesFile := filepath.Join(agentsDir, "rules")
		if err := os.WriteFile(rulesFile, []byte("test"), 0644); err != nil {
			t.Fatalf("create rules file: %v", err)
		}

		if internal.HasRulesDir(tmpDir) {
			t.Error("HasRulesDir should return false when rules is a file")
		}
	})
}

// Helper function to run git commands
func runGitCommand(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %v failed: %v", args, err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
