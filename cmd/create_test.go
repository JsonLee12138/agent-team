// cmd/create_test.go
package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
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

	// Verify task directories
	pendingDir := filepath.Join(wtPath, "agents", "teams", "backend", "tasks", "pending")
	if _, err := os.Stat(pendingDir); os.IsNotExist(err) {
		t.Error("tasks/pending not created")
	}
}

func TestRunCreateDuplicate(t *testing.T) {
	app, _ := initTestApp(t)
	app.RunCreate("dup")
	err := app.RunCreate("dup")
	if err == nil {
		t.Error("expected error for duplicate role")
	}
}
