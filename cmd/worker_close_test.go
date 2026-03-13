package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

type closeMockBackend struct {
	alivePanes  map[string]bool
	killedPanes []string
	killErr     error
}

func (m *closeMockBackend) PaneAlive(paneID string) bool {
	return m.alivePanes[paneID]
}

func (m *closeMockBackend) PaneSend(paneID string, text string) error {
	return nil
}

func (m *closeMockBackend) SpawnPane(cwd string, newWindow bool) (string, error) {
	return "", nil
}

func (m *closeMockBackend) KillPane(paneID string) error {
	if m.killErr != nil {
		return m.killErr
	}
	m.killedPanes = append(m.killedPanes, paneID)
	return nil
}

func (m *closeMockBackend) SetTitle(paneID string, title string) error {
	return nil
}

func (m *closeMockBackend) ActivatePane(paneID string) error {
	return nil
}

func TestRunWorkerCloseKillsLivePane(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &closeMockBackend{
		alivePanes: map[string]bool{"42": true},
	}
	app.Session = mock

	workerID := "dev-001"
	wtPath := filepath.Join(dir, app.WtBase, workerID)
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	cfg := &internal.WorkerConfig{
		WorkerID:         workerID,
		Role:             "dev",
		Provider:         "claude",
		DefaultModel:     "sonnet",
		PaneID:           "42",
		ControllerPaneID: "99",
	}
	if err := cfg.Save(internal.WorkerYAMLPath(wtPath)); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	if err := app.RunWorkerClose(workerID); err != nil {
		t.Fatalf("RunWorkerClose: %v", err)
	}

	if len(mock.killedPanes) != 1 || mock.killedPanes[0] != "42" {
		t.Fatalf("killed panes = %v, want [42]", mock.killedPanes)
	}

	reloaded, err := internal.LoadWorkerConfig(internal.WorkerYAMLPath(wtPath))
	if err != nil {
		t.Fatalf("LoadWorkerConfig: %v", err)
	}
	if reloaded.PaneID != "" {
		t.Fatalf("PaneID = %q, want empty", reloaded.PaneID)
	}
	// ControllerPaneID, Provider, DefaultModel must be preserved
	if reloaded.ControllerPaneID != "99" {
		t.Fatalf("ControllerPaneID = %q, want 99", reloaded.ControllerPaneID)
	}
	if reloaded.Provider != "claude" {
		t.Fatalf("Provider = %q, want claude", reloaded.Provider)
	}
	if reloaded.DefaultModel != "sonnet" {
		t.Fatalf("DefaultModel = %q, want sonnet", reloaded.DefaultModel)
	}
}

func TestRunWorkerCloseSucceedsWhenPaneAlreadyOffline(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &closeMockBackend{
		alivePanes: map[string]bool{"42": false},
	}
	app.Session = mock

	workerID := "dev-001"
	wtPath := filepath.Join(dir, app.WtBase, workerID)
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	cfg := &internal.WorkerConfig{
		WorkerID: workerID,
		Role:     "dev",
		Provider: "claude",
		PaneID:   "42",
	}
	if err := cfg.Save(internal.WorkerYAMLPath(wtPath)); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	if err := app.RunWorkerClose(workerID); err != nil {
		t.Fatalf("RunWorkerClose: %v", err)
	}

	// No kill should have been attempted
	if len(mock.killedPanes) != 0 {
		t.Fatalf("killed panes = %v, want none", mock.killedPanes)
	}

	reloaded, err := internal.LoadWorkerConfig(internal.WorkerYAMLPath(wtPath))
	if err != nil {
		t.Fatalf("LoadWorkerConfig: %v", err)
	}
	if reloaded.PaneID != "" {
		t.Fatalf("PaneID = %q, want empty (stale PaneID should be cleared)", reloaded.PaneID)
	}
}

func TestRunWorkerCloseSucceedsWhenPaneIDAlreadyEmpty(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &closeMockBackend{
		alivePanes: map[string]bool{},
	}
	app.Session = mock

	workerID := "dev-001"
	wtPath := filepath.Join(dir, app.WtBase, workerID)
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	cfg := &internal.WorkerConfig{
		WorkerID: workerID,
		Role:     "dev",
		Provider: "claude",
		PaneID:   "",
	}
	if err := cfg.Save(internal.WorkerYAMLPath(wtPath)); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	if err := app.RunWorkerClose(workerID); err != nil {
		t.Fatalf("RunWorkerClose: %v", err)
	}

	if len(mock.killedPanes) != 0 {
		t.Fatalf("killed panes = %v, want none", mock.killedPanes)
	}
}

func TestRunWorkerCloseReturnsErrorWhenKillFails(t *testing.T) {
	app, dir := initTestApp(t)
	mock := &closeMockBackend{
		alivePanes: map[string]bool{"42": true},
		killErr:    fmt.Errorf("pane locked"),
	}
	app.Session = mock

	workerID := "dev-001"
	wtPath := filepath.Join(dir, app.WtBase, workerID)
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	cfg := &internal.WorkerConfig{
		WorkerID: workerID,
		Role:     "dev",
		Provider: "claude",
		PaneID:   "42",
	}
	if err := cfg.Save(internal.WorkerYAMLPath(wtPath)); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	err := app.RunWorkerClose(workerID)
	if err == nil {
		t.Fatal("expected error when KillPane fails")
	}

	// PaneID should NOT be cleared when kill fails
	reloaded, loadErr := internal.LoadWorkerConfig(internal.WorkerYAMLPath(wtPath))
	if loadErr != nil {
		t.Fatalf("LoadWorkerConfig: %v", loadErr)
	}
	if reloaded.PaneID != "42" {
		t.Fatalf("PaneID = %q, want 42 (should not be cleared on kill failure)", reloaded.PaneID)
	}
}
