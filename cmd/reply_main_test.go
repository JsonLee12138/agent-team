// cmd/reply_main_test.go
package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

// MockBackend implements SessionBackend for testing.
type MockBackend struct {
	SpawnedID  string
	AlivePanes map[string]bool
	SentTexts  []string
}

func (m *MockBackend) PaneAlive(paneID string) bool {
	return m.AlivePanes[paneID]
}

func (m *MockBackend) PaneSend(paneID string, text string) error {
	m.SentTexts = append(m.SentTexts, text)
	return nil
}

func (m *MockBackend) SpawnPane(cwd string, _ bool) (string, error) {
	return m.SpawnedID, nil
}

func (m *MockBackend) KillPane(paneID string) error {
	return nil
}

func (m *MockBackend) SetTitle(paneID string, title string) error {
	return nil
}

func (m *MockBackend) ActivatePane(paneID string) error {
	return nil
}

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
		t.Fatalf("NewGitClient: %v", err)
	}

	app := &App{
		Git:     gc,
		Session: &MockBackend{AlivePanes: map[string]bool{}},
		WtBase:  ".worktrees",
	}
	return app, dir
}

func TestRunReplyMainV2(t *testing.T) {
	taskSetup = func(wtPath string) error { return nil }
	defer func() { taskSetup = defaultTaskSetup }()

	app, dir := initTestApp(t)
	mock := &MockBackend{
		SpawnedID:  "50",
		AlivePanes: map[string]bool{"50": true, "99": true},
	}
	app.Session = mock

	// Create role
	roleDir := filepath.Join(dir, ".agents", "teams", "dev")
	os.MkdirAll(roleDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# dev\n"), 0644)

	// Create worktree directory with worker.yaml
	wtPath := filepath.Join(dir, ".worktrees", "dev-001")
	os.MkdirAll(wtPath, 0755)
	wcfg := &internal.WorkerConfig{
		WorkerID:         "dev-001",
		Role:             "dev",
		Provider:         "claude",
		PaneID:           "50",
		ControllerPaneID: "99",
	}
	wcfg.Save(internal.WorkerYAMLPath(wtPath))

	// Override resolveWorktreeRoot to return the test worktree path
	origResolve := resolveWorktreeRoot
	resolveWorktreeRoot = func() (string, error) { return wtPath, nil }
	defer func() { resolveWorktreeRoot = origResolve }()

	err := app.RunReplyMain("What database should I use?")
	if err != nil {
		t.Fatalf("RunReplyMain: %v", err)
	}

	if len(mock.SentTexts) == 0 {
		t.Fatal("no text sent")
	}
	if !strings.Contains(mock.SentTexts[0], "[Worker: dev-001]") {
		t.Errorf("sent = %q, want [Worker: dev-001] prefix", mock.SentTexts[0])
	}
	if !strings.Contains(mock.SentTexts[0], "What database should I use?") {
		t.Errorf("sent = %q, want message content", mock.SentTexts[0])
	}
}

func TestRunReplyMainNoController(t *testing.T) {
	app, dir := initTestApp(t)

	// Create worktree directory with worker.yaml (no controller pane)
	wtPath := filepath.Join(dir, ".worktrees", "solo-001")
	os.MkdirAll(wtPath, 0755)
	wcfg := &internal.WorkerConfig{WorkerID: "solo-001", Role: "solo", Provider: "claude"}
	wcfg.Save(internal.WorkerYAMLPath(wtPath))

	// Override resolveWorktreeRoot to return the test worktree path
	origResolve := resolveWorktreeRoot
	resolveWorktreeRoot = func() (string, error) { return wtPath, nil }
	defer func() { resolveWorktreeRoot = origResolve }()

	err := app.RunReplyMain("hello")
	if err == nil {
		t.Error("expected error when no controller pane ID")
	}
}
