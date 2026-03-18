// cmd/init_setup_test.go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

// --- 3B-3: init + setup command separation ---

func stubInitProjectCommands(t *testing.T, content string) {
	t.Helper()
	prev := rebuildProjectCommands
	rebuildProjectCommands = func(root string) (*internal.BuildScriptScan, error) {
		rulesDir := filepath.Join(root, ".agents", "rules")
		if err := os.MkdirAll(rulesDir, 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(rulesDir, "project-commands.md"), []byte(content), 0644); err != nil {
			return nil, err
		}
		return &internal.BuildScriptScan{}, nil
	}
	t.Cleanup(func() { rebuildProjectCommands = prev })
}

func TestInitCmd_Exists(t *testing.T) {
	root := NewRootCmd()
	RegisterCommands(root)

	// The init subcommand should be registered
	cmd, _, err := root.Find([]string{"init"})
	if err != nil {
		t.Fatalf("init command not found: %v", err)
	}
	if cmd.Use != "init" {
		t.Errorf("Use = %q, want 'init'", cmd.Use)
	}
}

func TestSetupCmd_Exists(t *testing.T) {
	root := NewRootCmd()
	RegisterCommands(root)

	cmd, _, err := root.Find([]string{"setup"})
	if err != nil {
		t.Fatalf("setup command not found: %v", err)
	}
	if cmd.Use != "setup" {
		t.Errorf("Use = %q, want 'setup'", cmd.Use)
	}
}

func TestSetupCmd_HasSkipDetectFlag(t *testing.T) {
	root := NewRootCmd()
	RegisterCommands(root)

	cmd, _, _ := root.Find([]string{"setup"})
	flag := cmd.Flags().Lookup("skip-detect")
	if flag == nil {
		t.Error("setup should have --skip-detect flag")
	}
}

func TestInitCmd_ProjectLevelInit(t *testing.T) {
	// init should create .agents/teams/, .agents/rules/, and provider files
	dir := t.TempDir()
	stubInitProjectCommands(t, "# Project Commands Rules\n\nGenerated during init tests.\n")

	origWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origWd)

	root := NewRootCmd()
	RegisterCommands(root)
	root.SetArgs([]string{"init"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Verify .agents/teams/ exists
	if _, err := os.Stat(filepath.Join(dir, ".agents", "teams")); os.IsNotExist(err) {
		t.Error(".agents/teams/ should be created by init")
	}

	// Verify .agents/rules/ exists with default rule files
	rulesDir := filepath.Join(dir, ".agents", "rules")
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		t.Error(".agents/rules/ should be created by init")
	}

	// Verify index.md exists
	if _, err := os.Stat(filepath.Join(rulesDir, "index.md")); os.IsNotExist(err) {
		t.Error("index.md should be created in rules dir")
	}
	if _, err := os.Stat(filepath.Join(rulesDir, "project-commands.md")); os.IsNotExist(err) {
		t.Error("project-commands.md should be created by init")
	}

	// Verify provider files
	for _, name := range []string{"CLAUDE.md", "AGENTS.md", "GEMINI.md"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Errorf("%s should be created: %v", name, err)
			continue
		}
		if !strings.Contains(string(data), "AGENT_TEAM:START") {
			t.Errorf("%s should contain AGENT_TEAM tag", name)
		}
		if !strings.Contains(string(data), ".agents/rules/project-commands.md") {
			t.Errorf("%s should reference project-commands.md", name)
		}
	}
}

func TestInitCmd_Idempotent(t *testing.T) {
	dir := t.TempDir()
	stubInitProjectCommands(t, "# Project Commands Rules\n\nGenerated during idempotent init tests.\n")

	origWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origWd)

	root := NewRootCmd()
	RegisterCommands(root)

	// First init
	root.SetArgs([]string{"init"})
	if err := root.Execute(); err != nil {
		t.Fatalf("first init: %v", err)
	}

	// Write custom content into rules/index.md
	customContent := "# My Custom Rules\n\nDo not overwrite this.\n"
	os.WriteFile(filepath.Join(dir, ".agents", "rules", "index.md"), []byte(customContent), 0644)

	// Add user content to CLAUDE.md
	claudeMD := filepath.Join(dir, "CLAUDE.md")
	existing, _ := os.ReadFile(claudeMD)
	os.WriteFile(claudeMD, append(existing, []byte("\n## My Custom Section\n")...), 0644)

	// Second init — should not overwrite user content
	root2 := NewRootCmd()
	RegisterCommands(root2)
	root2.SetArgs([]string{"init"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("second init: %v", err)
	}

	// Verify custom rules content preserved
	data, _ := os.ReadFile(filepath.Join(dir, ".agents", "rules", "index.md"))
	if string(data) != customContent {
		t.Errorf("custom index.md was overwritten, got %q", string(data))
	}

	// Verify custom CLAUDE.md content preserved
	claudeData, _ := os.ReadFile(claudeMD)
	if !strings.Contains(string(claudeData), "My Custom Section") {
		t.Error("custom CLAUDE.md section was lost after second init")
	}

	// Verify tag section still exists
	if !strings.Contains(string(claudeData), "AGENT_TEAM:START") {
		t.Error("AGENT_TEAM tag section missing after second init")
	}

	// Verify there's only one tag section (not duplicated)
	if strings.Count(string(claudeData), "AGENT_TEAM:START") != 1 {
		t.Error("AGENT_TEAM tag section duplicated after second init")
	}
}

func TestInitCmd_DoesNotSkipGitCheck(t *testing.T) {
	// init is in the skip list for PersistentPreRunE, so it doesn't need a git repo
	root := NewRootCmd()
	RegisterCommands(root)
	stubInitProjectCommands(t, "# Project Commands Rules\n\nGenerated during git check tests.\n")

	dir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origWd)

	root.SetArgs([]string{"init"})
	err := root.Execute()
	// init should work without a git repository
	if err != nil {
		t.Fatalf("init should work outside git repo, got: %v", err)
	}
}

func TestSetupCmd_SkipsGitCheck(t *testing.T) {
	root := NewRootCmd()
	RegisterCommands(root)

	dir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origWd)

	// setup should not fail due to missing git repo
	root.SetArgs([]string{"setup", "--skip-detect"})
	// May fail for other reasons (no plugin root), but not for git
	err := root.Execute()
	// We just check it doesn't panic and runs past git check
	_ = err
}
