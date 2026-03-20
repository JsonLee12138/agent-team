// internal/git.go
package internal

import (
	"fmt"
	"os/exec"
	"path/filepath"
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

func ResolveProjectRootFromWorktree(worktreeRoot string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	cmd.Dir = worktreeRoot
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolve git common dir for %s: %w", worktreeRoot, err)
	}
	commonDir := strings.TrimSpace(string(out))
	if commonDir == "" || commonDir == ".git" || commonDir == "./.git" {
		return worktreeRoot, nil
	}
	if filepath.IsAbs(commonDir) {
		return filepath.Clean(filepath.Join(commonDir, "..")), nil
	}
	return filepath.Clean(filepath.Join(worktreeRoot, commonDir, "..")), nil
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

// BranchExists checks if a git branch exists.
func (g *GitClient) BranchExists(branch string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", branch)
	cmd.Dir = g.root
	return cmd.Run() == nil
}

func (g *GitClient) WorktreeAdd(path, branch string) error {
	var cmd *exec.Cmd
	if g.BranchExists(branch) {
		// Branch already exists (e.g. leftover from incomplete cleanup), reuse it
		cmd = exec.Command("git", "worktree", "add", path, branch)
	} else {
		cmd = exec.Command("git", "worktree", "add", path, "-b", branch)
	}
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

func (g *GitClient) RebaseWorktree(wtPath, onto string) error {
	cmd := exec.Command("git", "-C", wtPath, "rebase", onto)
	cmd.Dir = g.root
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("rebase worktree %s onto %s: %s (%w)", wtPath, onto, out, err)
	}
	return nil
}

func (g *GitClient) DeleteBranch(branch string) error {
	cmd := exec.Command("git", "branch", "-D", branch)
	cmd.Dir = g.root
	cmd.CombinedOutput() // ignore error — branch may not exist
	return nil
}
