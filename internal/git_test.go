// internal/git_test.go
package internal

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func initTestRepo(t *testing.T) string {
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
	return dir
}

func TestNewGitClient(t *testing.T) {
	dir := initTestRepo(t)
	gc, err := NewGitClient(dir)
	if err != nil {
		t.Fatalf("NewGitClient: %v", err)
	}
	// git rev-parse resolves symlinks; normalize expected path for macOS /private/var
	want, _ := filepath.EvalSymlinks(dir)
	if gc.Root() != want {
		t.Errorf("Root() = %q, want %q", gc.Root(), want)
	}
}

func TestGitClientCurrentBranch(t *testing.T) {
	dir := initTestRepo(t)
	gc, _ := NewGitClient(dir)
	branch, err := gc.CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	// git init creates "main" or "master" depending on config
	if branch == "" {
		t.Error("CurrentBranch returned empty string")
	}
}

func TestGitClientWorktreeAddRemove(t *testing.T) {
	dir := initTestRepo(t)
	gc, _ := NewGitClient(dir)

	wtPath := dir + "/.worktrees/test-role"
	if err := gc.WorktreeAdd(wtPath, "team/test-role"); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}

	if err := gc.WorktreeRemove(wtPath); err != nil {
		t.Fatalf("WorktreeRemove: %v", err)
	}
}
