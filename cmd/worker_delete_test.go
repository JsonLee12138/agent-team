package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

type deleteMockBackend struct {
	alivePanes  map[string]bool
	killedPanes []string
	killErr     error
}

func (m *deleteMockBackend) PaneAlive(paneID string) bool {
	return m.alivePanes[paneID]
}

func (m *deleteMockBackend) PaneSend(paneID string, text string) error {
	return nil
}

func (m *deleteMockBackend) SpawnPane(cwd string, newWindow bool) (string, error) {
	return "", nil
}

func (m *deleteMockBackend) KillPane(paneID string) error {
	if m.killErr != nil {
		return m.killErr
	}
	m.killedPanes = append(m.killedPanes, paneID)
	return nil
}

func (m *deleteMockBackend) SetTitle(paneID string, title string) error {
	return nil
}

func (m *deleteMockBackend) ActivatePane(paneID string) error {
	return nil
}

func TestRunWorkerDeleteKillsAlivePane(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &deleteMockBackend{
		alivePanes: map[string]bool{"50": true},
	}
	app.Session = mock

	workerID := "dev-001"
	wtPath := filepath.Join(dir, app.WtBase, workerID)
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}

	cfg := &internal.WorkerConfig{
		WorkerID: "dev-001",
		Role:     "dev",
		Provider: "claude",
		PaneID:   "50",
	}
	if err := cfg.Save(internal.WorkerYAMLPath(wtPath)); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	if err := app.RunWorkerDelete(workerID); err != nil {
		t.Fatalf("RunWorkerDelete: %v", err)
	}

	if len(mock.killedPanes) != 1 || mock.killedPanes[0] != "50" {
		t.Fatalf("killed panes = %v, want [50]", mock.killedPanes)
	}
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Fatalf("worktree still exists after delete: %s", wtPath)
	}
}

func TestRunWorkerDeleteSkipsKillWhenPaneOffline(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &deleteMockBackend{
		alivePanes: map[string]bool{"50": false},
	}
	app.Session = mock

	workerID := "dev-001"
	wtPath := filepath.Join(dir, app.WtBase, workerID)
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}

	cfg := &internal.WorkerConfig{
		WorkerID: "dev-001",
		Role:     "dev",
		Provider: "claude",
		PaneID:   "50",
	}
	if err := cfg.Save(internal.WorkerYAMLPath(wtPath)); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	if err := app.RunWorkerDelete(workerID); err != nil {
		t.Fatalf("RunWorkerDelete: %v", err)
	}

	if len(mock.killedPanes) != 0 {
		t.Fatalf("killed panes = %v, want none", mock.killedPanes)
	}
}

func TestRunWorkerDeleteStopsWhenCloseFails(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &deleteMockBackend{
		alivePanes: map[string]bool{"50": true},
		killErr:    fmt.Errorf("pane locked"),
	}
	app.Session = mock

	workerID := "dev-001"
	wtPath := filepath.Join(dir, app.WtBase, workerID)
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}

	cfg := &internal.WorkerConfig{
		WorkerID: "dev-001",
		Role:     "dev",
		Provider: "claude",
		PaneID:   "50",
	}
	if err := cfg.Save(internal.WorkerYAMLPath(wtPath)); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	err := app.RunWorkerDelete(workerID)
	if err == nil {
		t.Fatal("expected delete to fail when close fails for live pane")
	}

	// worktree should still exist since delete was aborted
	if _, statErr := os.Stat(wtPath); os.IsNotExist(statErr) {
		t.Fatal("worktree should not be removed when close fails")
	}
}
