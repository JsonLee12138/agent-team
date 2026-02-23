// internal/git.go
package internal

import (
	"fmt"
	"os/exec"
	"strings"
)

type GitClient struct {
	root string
}

func NewGitClient(dir string) (*GitClient, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}
	return &GitClient{root: strings.TrimSpace(string(out))}, nil
}

func (g *GitClient) Root() string {
	return g.root
}

func (g *GitClient) CurrentBranch() (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = g.root
	out, err := cmd.Output()
	if err != nil {
		return "main", nil
	}
	return strings.TrimSpace(string(out)), nil
}

func (g *GitClient) WorktreeAdd(path, branch string) error {
	cmd := exec.Command("git", "worktree", "add", path, "-b", branch)
	cmd.Dir = g.root
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("worktree add: %s (%w)", out, err)
	}
	return nil
}

func (g *GitClient) WorktreeRemove(path string) error {
	cmd := exec.Command("git", "worktree", "remove", path, "--force")
	cmd.Dir = g.root
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("worktree remove: %s (%w)", out, err)
	}
	return nil
}

func (g *GitClient) Merge(branch, message string) error {
	cmd := exec.Command("git", "merge", branch, "--no-ff", "-m", message)
	cmd.Dir = g.root
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("merge: %s (%w)", out, err)
	}
	return nil
}

func (g *GitClient) DeleteBranch(branch string) error {
	cmd := exec.Command("git", "branch", "-D", branch)
	cmd.Dir = g.root
	cmd.CombinedOutput() // ignore error — branch may not exist
	return nil
}
