package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

var resolveGitTopLevel = func() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func samePath(a, b string) bool {
	aEval, errA := filepath.EvalSymlinks(a)
	bEval, errB := filepath.EvalSymlinks(b)
	if errA == nil {
		a = aEval
	}
	if errB == nil {
		b = bEval
	}
	return filepath.Clean(a) == filepath.Clean(b)
}
