// cmd/init_setup_test.go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

// --- init command coverage ---

func stubInitProjectCommands(t *testing.T, content string) {
	t.Helper()
	prev := rebuildProjectRules
	rebuildProjectRules = func(root string) (*internal.BuildScriptScan, error) {
		rulesDir := filepath.Join(root, ".agent-team", "rules", "project")
		if err := os.MkdirAll(rulesDir, 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(rulesDir, "project-commands.md"), []byte(content), 0644); err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(rulesDir, "project-constraints.md"), []byte("# Project Constraints\n\nGenerated during init tests.\n"), 0644); err != nil {
			return nil, err
		}
		return &internal.BuildScriptScan{}, nil
	}
	t.Cleanup(func() { rebuildProjectRules = prev })
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

func TestSetupCmd_NotRegistered(t *testing.T) {
	root := NewRootCmd()
	RegisterCommands(root)

	cmd, _, err := root.Find([]string{"setup"})
	if err == nil {
		t.Fatalf("expected setup to be removed, got command %q", cmd.CommandPath())
	}
}

func TestInitCmd_ProjectLevelInit(t *testing.T) {
	// init should create .agent-team/teams/, .agent-team/rules/, and provider files
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

	// Verify .agent-team/teams/ exists
	if _, err := os.Stat(filepath.Join(dir, ".agent-team", "teams")); os.IsNotExist(err) {
		t.Error(".agent-team/teams/ should be created by init")
	}
	for _, rel := range []string{
		filepath.Join(".agent-team", "planning", "roadmaps"),
		filepath.Join(".agent-team", "planning", "milestones"),
		filepath.Join(".agent-team", "planning", "phases"),
		filepath.Join(".agent-team", "archive", "roadmaps"),
		filepath.Join(".agent-team", "archive", "milestones"),
		filepath.Join(".agent-team", "archive", "phases"),
		filepath.Join(".agent-team", "deprecated", "roadmaps"),
		filepath.Join(".agent-team", "deprecated", "milestones"),
		filepath.Join(".agent-team", "deprecated", "phases"),
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); os.IsNotExist(err) {
			t.Errorf("%s should be created by init", rel)
		}
	}

	// Verify .agent-team/rules/ exists with default rule files
	rulesDir := filepath.Join(dir, ".agent-team", "rules")
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		t.Error(".agent-team/rules/ should be created by init")
	}

	// Verify index.md and representative rule files exist
	for _, name := range []string{"index.md", "core/agent-team-commands.md", "core/merge-workflow.md", "project/project-commands.md"} {
		if _, err := os.Stat(filepath.Join(rulesDir, name)); os.IsNotExist(err) {
			t.Errorf("%s should be created in rules dir", name)
		}
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
		content := string(data)
		if !strings.Contains(content, ".agent-team/rules/project/") {
			t.Errorf("%s should reference .agent-team/rules/project/", name)
		}
		if strings.Contains(content, "MUST call `/compact`") {
			t.Errorf("%s should not require /compact", name)
		}
		for _, needle := range []string{"context-cleanup", "index-first recovery"} {
			if !strings.Contains(content, needle) {
				t.Errorf("%s should reference %s", name, needle)
			}
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

	// Write custom content into a core rule file
	customContent := "# My Custom Rules\n\nDo not overwrite this.\n"
	os.WriteFile(filepath.Join(dir, ".agent-team", "rules", "core", "debugging.md"), []byte(customContent), 0644)

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
	data, _ := os.ReadFile(filepath.Join(dir, ".agent-team", "rules", "core", "debugging.md"))
	if string(data) != customContent {
		t.Errorf("custom core rule was overwritten, got %q", string(data))
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

