package cmd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

// TestPersistentPreRunE_UninitializedProject tests the initialization check
// in PersistentPreRunE when .agent-team/rules/ does not exist.
func TestPersistentPreRunE_UninitializedProject(t *testing.T) {
	runPreRun := func(t *testing.T, branch string) error {
		t.Helper()

		tmpDir := t.TempDir()
		runGitCommand(t, tmpDir, "init")
		runGitCommand(t, tmpDir, "symbolic-ref", "HEAD", "refs/heads/"+branch)

		origCwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd: %v", err)
		}
		defer os.Chdir(origCwd) // nolint:errcheck
		os.Chdir(tmpDir)        // nolint:errcheck

		t.Setenv("AGENT_TEAM_NONINTERACTIVE", "1")

		rootCmd := NewRootCmd()
		workerCmd := &cobra.Command{
			Use: "worker",
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}
		workerCmd.SetContext(context.Background())
		rootCmd.AddCommand(workerCmd)

		if rootCmd.PersistentPreRunE == nil {
			t.Fatal("PersistentPreRunE is not set")
		}

		return rootCmd.PersistentPreRunE(workerCmd, []string{})
	}

	t.Run("main branch requires initialization", func(t *testing.T) {
		err := runPreRun(t, "main")
		if err == nil {
			t.Fatal("Expected error when .agent-team/rules/ not found on main branch")
		}
		if !contains(err.Error(), ".agent-team/rules/ not found") {
			t.Errorf("Expected initialization error, got: %v", err)
		}
	})

	t.Run("team branch bypasses initialization", func(t *testing.T) {
		err := runPreRun(t, "team/test-001")
		if err != nil {
			t.Fatalf("Expected no error on team/* branch, got: %v", err)
		}
	})

	t.Run("non-team branch still requires initialization", func(t *testing.T) {
		err := runPreRun(t, "feature/foo")
		if err == nil {
			t.Fatal("Expected error when .agent-team/rules/ not found on non-team branch")
		}
		if !contains(err.Error(), ".agent-team/rules/ not found") {
			t.Errorf("Expected initialization error, got: %v", err)
		}
	})
}

// TestPersistentPreRunE_InitializedProject tests that Initialized project
// passes PersistentPreRunE without error.
func TestPersistentPreRunE_InitializedProject(t *testing.T) {
	t.Run("passes when .agent-team/rules/ exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize git repo
		runGitCommand(t, tmpDir, "init")

		// Create .agent-team/rules/ directory
		rulesDir := filepath.Join(tmpDir, ".agent-team", "rules")
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
			if contains(err.Error(), ".agent-team/rules/") {
				t.Errorf("Should pass initialization check, got: %v", err)
			}
		}
	})
}

// TestHasRulesDir_Integration tests HasRulesDir helper in various scenarios.
func TestHasRulesDir_Integration(t *testing.T) {
	t.Run("returns true when .agent-team/rules/ exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		rulesDir := filepath.Join(tmpDir, ".agent-team", "rules")
		if err := os.MkdirAll(rulesDir, 0755); err != nil {
			t.Fatalf("create rules dir: %v", err)
		}

		if !internal.HasRulesDir(tmpDir) {
			t.Error("HasRulesDir should return true when .agent-team/rules/ exists")
		}
	})

	t.Run("returns false when .agent-team/ exists but rules/ does not", func(t *testing.T) {
		tmpDir := t.TempDir()
		agentsDir := filepath.Join(tmpDir, ".agent-team")
		if err := os.MkdirAll(agentsDir, 0755); err != nil {
			t.Fatalf("create agents dir: %v", err)
		}

		if internal.HasRulesDir(tmpDir) {
			t.Error("HasRulesDir should return false when rules/ does not exist")
		}
	})

	t.Run("returns false when .agent-team/ does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		if internal.HasRulesDir(tmpDir) {
			t.Error("HasRulesDir should return false when .agent-team/ does not exist")
		}
	})

	t.Run("returns false when rules is a file not a directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		agentsDir := filepath.Join(tmpDir, ".agent-team")
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
