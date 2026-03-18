package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRulesSyncCmdHasRebuildFlag(t *testing.T) {
	cmd := newRulesSyncCmd()
	if flag := cmd.Flags().Lookup("rebuild"); flag != nil {
		t.Fatal("sync command should not expose --rebuild")
	}
}

func TestRulesSyncRefreshesStaticRulesAndProjectCommands(t *testing.T) {
	_, dir := initTestApp(t)
	prevRebuild := rebuildProjectCommands
	rebuildProjectCommands = func(root string) (*internal.BuildScriptScan, error) {
		path := filepath.Join(root, ".agents", "rules", "project-commands.md")
		if err := os.WriteFile(path, []byte("# Project Commands Rules\n\nSynced from tests.\n"), 0644); err != nil {
			return nil, err
		}
		return &internal.BuildScriptScan{}, nil
	}
	t.Cleanup(func() { rebuildProjectCommands = prevRebuild })

	if err := os.MkdirAll(filepath.Join(dir, ".agents", "rules"), 0755); err != nil {
		t.Fatalf("mkdir rules dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".agents", "rules", "index.md"), []byte("# stale\n"), 0644); err != nil {
		t.Fatalf("write stale index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".agents", "rules", "build-verification.md"), []byte("legacy\n"), 0644); err != nil {
		t.Fatalf("write legacy build-verification.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:\n\tgo build ./...\n\ntest:\n\tgo test ./...\n"), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer os.Chdir(origWd)

	root := NewRootCmd()
	RegisterCommands(root)
	root.SetArgs([]string{"rules", "sync"})
	if err := root.Execute(); err != nil {
		t.Fatalf("rules sync: %v", err)
	}

	indexData, err := os.ReadFile(filepath.Join(dir, ".agents", "rules", "index.md"))
	if err != nil {
		t.Fatalf("read index.md: %v", err)
	}
	if !strings.Contains(string(indexData), "project-commands.md") {
		t.Fatalf("index.md = %q, want project-commands reference", string(indexData))
	}

	data, err := os.ReadFile(filepath.Join(dir, ".agents", "rules", "project-commands.md"))
	if err != nil {
		t.Fatalf("read project-commands.md: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "Synced from tests.") {
		t.Fatalf("project-commands.md = %q, want regenerated content", content)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "rules", "build-verification.md")); !os.IsNotExist(err) {
		t.Fatalf("legacy build-verification.md should be removed, err=%v", err)
	}

	agentData, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(agentData), ".agents/rules/project-commands.md") {
		t.Fatalf("AGENTS.md should reference project-commands.md, got:\n%s", string(agentData))
	}
}
