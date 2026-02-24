// cmd/create_test.go
package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

func initTestApp(t *testing.T) (*App, string) {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s (%v)", args, out, err)
		}
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	run("commit", "--allow-empty", "-m", "init")

	gc, err := internal.NewGitClient(dir)
	if err != nil {
		t.Fatal(err)
	}

	return &App{
		Git:     gc,
		Session: &MockBackend{},
		WtBase:  ".worktrees",
	}, dir
}

// MockBackend implements SessionBackend for testing
type MockBackend struct {
	AlivePanes map[string]bool
	SentTexts  []string
	SpawnedID  string
}

func (m *MockBackend) PaneAlive(id string) bool          { return m.AlivePanes[id] }
func (m *MockBackend) PaneSend(_ string, t string) error { m.SentTexts = append(m.SentTexts, t); return nil }
func (m *MockBackend) SpawnPane(_ string) (string, error) { return m.SpawnedID, nil }
func (m *MockBackend) KillPane(_ string) error            { return nil }
func (m *MockBackend) SetTitle(_, _ string) error         { return nil }
func (m *MockBackend) ActivatePane(_ string) error        { return nil }

func TestRunCreate(t *testing.T) {
	// Track whether prompt.md exists when openspec runs
	promptExistedDuringOpenSpec := false
	openSpecSetup = func(wtPath string) error {
		// Check if prompt.md already exists at this point
		name := filepath.Base(wtPath)
		promptPath := filepath.Join(wtPath, "agents", "teams", name, "prompt.md")
		if _, err := os.Stat(promptPath); err == nil {
			promptExistedDuringOpenSpec = true
		}
		return nil
	}
	defer func() { openSpecSetup = defaultOpenSpecSetup }()

	app, dir := initTestApp(t)

	if err := app.RunCreate("backend"); err != nil {
		t.Fatalf("RunCreate: %v", err)
	}

	// Verify worktree directory
	wtPath := filepath.Join(dir, ".worktrees", "backend")
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Error("worktree directory not created")
	}

	// Verify config.yaml
	configPath := filepath.Join(wtPath, "agents", "teams", "backend", "config.yaml")
	cfg, err := internal.LoadRoleConfig(configPath)
	if err != nil {
		t.Fatalf("LoadRoleConfig: %v", err)
	}
	if cfg.Name != "backend" {
		t.Errorf("config.Name = %q, want backend", cfg.Name)
	}
	if cfg.DefaultProvider != "claude" {
		t.Errorf("config.DefaultProvider = %q, want claude", cfg.DefaultProvider)
	}

	// Verify prompt.md
	promptPath := filepath.Join(wtPath, "agents", "teams", "backend", "prompt.md")
	if _, err := os.Stat(promptPath); os.IsNotExist(err) {
		t.Error("prompt.md not created")
	}

	// Verify prompt.md existed before openspec init ran
	if !promptExistedDuringOpenSpec {
		t.Error("prompt.md should exist before openspec init runs")
	}

	// Verify CLAUDE.md contains AGENT_TEAM tags
	claudeMD := filepath.Join(wtPath, "CLAUDE.md")
	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("CLAUDE.md not found: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "<!-- AGENT_TEAM:START -->") {
		t.Error("CLAUDE.md should contain AGENT_TEAM start marker")
	}
	if !strings.Contains(content, "<!-- AGENT_TEAM:END -->") {
		t.Error("CLAUDE.md should contain AGENT_TEAM end marker")
	}

	// Verify AGENTS.md contains AGENT_TEAM tags
	agentsMD := filepath.Join(wtPath, "AGENTS.md")
	agentsData, err := os.ReadFile(agentsMD)
	if err != nil {
		t.Fatalf("AGENTS.md not found: %v", err)
	}
	agentsContent := string(agentsData)
	if !strings.Contains(agentsContent, "<!-- AGENT_TEAM:START -->") {
		t.Error("AGENTS.md should contain AGENT_TEAM start marker")
	}

	// Verify old task dirs are NOT created
	pendingDir := filepath.Join(wtPath, "agents", "teams", "backend", "tasks", "pending")
	if _, err := os.Stat(pendingDir); err == nil {
		t.Error("tasks/pending should no longer be created")
	}
}

func TestRunCreateDuplicate(t *testing.T) {
	openSpecSetup = func(wtPath string) error { return nil }
	defer func() { openSpecSetup = defaultOpenSpecSetup }()

	app, _ := initTestApp(t)
	app.RunCreate("dup")
	err := app.RunCreate("dup")
	if err == nil {
		t.Error("expected error for duplicate role")
	}
}
