package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()
	w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy: %v", err)
	}
	return buf.String()
}

func TestRunContextCleanupControllerUsesRulesIndex(t *testing.T) {
	app, dir := initTestApp(t)
	origResolve := resolveGitTopLevel
	resolveGitTopLevel = func() (string, error) { return dir, nil }
	defer func() { resolveGitTopLevel = origResolve }()

	out := captureStdout(t, func() {
		if err := app.RunContextCleanup(""); err != nil {
			t.Fatalf("RunContextCleanup: %v", err)
		}
	})

	if !strings.Contains(out, filepath.Join(dir, ".agent-team", "rules", "index.md")) {
		t.Fatalf("output should mention rules index first, got:\n%s", out)
	}
	if !strings.Contains(out, "not context compression") {
		t.Fatalf("output should state non-compression semantics, got:\n%s", out)
	}
}

func TestRunContextCleanupWorkerUsesWorkerAndTaskArtifacts(t *testing.T) {
	app, dir := initTestApp(t)
	wtPath := filepath.Join(dir, ".worktrees", "dev-001")
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	record, err := internal.CreateTaskPackage(dir, "Implement feature", "dev", "", time.Now().UTC())
	if err != nil {
		t.Fatalf("CreateTaskPackage: %v", err)
	}
	cfg := &internal.WorkerConfig{WorkerID: "dev-001", Role: "dev", Provider: "claude", TaskID: record.TaskID, TaskPath: record.TaskPath}
	if err := cfg.Save(internal.WorkerYAMLPath(wtPath)); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	origResolve := resolveGitTopLevel
	resolveGitTopLevel = func() (string, error) { return wtPath, nil }
	defer func() { resolveGitTopLevel = origResolve }()

	out := captureStdout(t, func() {
		if err := app.RunContextCleanup(""); err != nil {
			t.Fatalf("RunContextCleanup: %v", err)
		}
	})

	if !strings.Contains(out, internal.WorkerYAMLPath(wtPath)) {
		t.Fatalf("output should mention worker.yaml first, got:\n%s", out)
	}
	if !strings.Contains(out, internal.TaskYAMLPath(dir, record.TaskID)) {
		t.Fatalf("output should mention task.yaml, got:\n%s", out)
	}
	if !strings.Contains(out, internal.TaskContextPath(dir, record.TaskID)) {
		t.Fatalf("output should mention context.md as needed, got:\n%s", out)
	}
	if !strings.Contains(out, internal.TaskVerificationPath(dir, record.TaskID)) {
		t.Fatalf("output should mention verification.md as needed, got:\n%s", out)
	}
	if !strings.Contains(out, "not context compression") {
		t.Fatalf("output should state non-compression semantics, got:\n%s", out)
	}
}

func TestRunContextCleanupFromControllerForWorkerUsesBoundWorker(t *testing.T) {
	app, dir := initTestApp(t)
	record, err := internal.CreateTaskPackage(dir, "Implement feature", "dev", "", time.Now().UTC())
	if err != nil {
		t.Fatalf("CreateTaskPackage: %v", err)
	}
	wtPath := filepath.Join(dir, ".worktrees", "dev-001")
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	cfg := &internal.WorkerConfig{WorkerID: "dev-001", Role: "dev", Provider: "claude", TaskID: record.TaskID, TaskPath: record.TaskPath}
	if err := cfg.Save(internal.WorkerYAMLPath(wtPath)); err != nil {
		t.Fatalf("save worker config: %v", err)
	}

	origResolve := resolveGitTopLevel
	resolveGitTopLevel = func() (string, error) { return dir, nil }
	defer func() { resolveGitTopLevel = origResolve }()

	out := captureStdout(t, func() {
		if err := app.RunContextCleanup("dev-001"); err != nil {
			t.Fatalf("RunContextCleanup: %v", err)
		}
	})

	if !strings.Contains(out, internal.WorkerYAMLPath(wtPath)) {
		t.Fatalf("output should mention worker yaml path, got:\n%s", out)
	}
	if !strings.Contains(out, internal.TaskYAMLPath(dir, record.TaskID)) {
		t.Fatalf("output should mention task yaml path, got:\n%s", out)
	}
	if !strings.Contains(out, internal.TaskVerificationPath(dir, record.TaskID)) {
		t.Fatalf("output should mention verification path, got:\n%s", out)
	}
}

func TestRegisterCommandsAddsContextCleanup(t *testing.T) {
	root := NewRootCmd()
	RegisterCommands(root)
	if _, _, err := root.Find([]string{"context-cleanup"}); err != nil {
		t.Fatalf("context-cleanup command should be registered: %v", err)
	}
}
