// internal/role_test.go
package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindWtBase(t *testing.T) {
	dir := t.TempDir()

	// no worktrees dir exists → default ".worktrees"
	if got := FindWtBase(dir); got != ".worktrees" {
		t.Errorf("FindWtBase(empty) = %q, want .worktrees", got)
	}

	// create .worktrees
	os.Mkdir(filepath.Join(dir, ".worktrees"), 0755)
	if got := FindWtBase(dir); got != ".worktrees" {
		t.Errorf("FindWtBase(.worktrees) = %q, want .worktrees", got)
	}
}

func TestListRoles(t *testing.T) {
	dir := t.TempDir()
	wtBase := ".worktrees"

	// empty → no roles
	roles := ListRoles(dir, wtBase)
	if len(roles) != 0 {
		t.Errorf("ListRoles(empty) = %v, want empty", roles)
	}

	// create a role
	teamsDir := filepath.Join(dir, wtBase, "backend", "agents", "teams", "backend")
	os.MkdirAll(teamsDir, 0755)
	os.WriteFile(filepath.Join(teamsDir, "config.yaml"), []byte("name: backend\n"), 0644)

	roles = ListRoles(dir, wtBase)
	if len(roles) != 1 || roles[0] != "backend" {
		t.Errorf("ListRoles = %v, want [backend]", roles)
	}
}

func TestBuildLaunchCmd(t *testing.T) {
	tests := []struct {
		provider, model, want string
	}{
		{"claude", "", "claude --dangerously-skip-permissions"},
		{"codex", "gpt-5", "codex --dangerously-bypass-approvals-and-sandbox --model gpt-5"},
		{"opencode", "", "opencode"},
		{"", "", "claude --dangerously-skip-permissions"},
	}
	for _, tt := range tests {
		got := BuildLaunchCmd(tt.provider, tt.model)
		if got != tt.want {
			t.Errorf("BuildLaunchCmd(%q, %q) = %q, want %q", tt.provider, tt.model, got, tt.want)
		}
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"Fix the login bug", "fix-the-login-bug"},
		{"Hello, World!!!", "hello-world"},
		{"", "task"},
	}
	for _, tt := range tests {
		got := Slugify(tt.input, 50)
		if got != tt.want {
			t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGenerateClaudeMD(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, ".worktrees", "dev")
	teamsDir := filepath.Join(wtPath, "agents", "teams", "dev")
	os.MkdirAll(teamsDir, 0755)

	promptContent := "# Role: dev\n\nA developer role.\n"
	os.WriteFile(filepath.Join(teamsDir, "prompt.md"), []byte(promptContent), 0644)

	err := GenerateClaudeMD(wtPath, "dev", dir)
	if err != nil {
		t.Fatalf("GenerateClaudeMD: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(wtPath, "CLAUDE.md"))
	content := string(data)

	if !strings.HasPrefix(content, "# Role: dev") {
		t.Error("CLAUDE.md should start with prompt.md content")
	}
	if !strings.Contains(content, "Development Environment") {
		t.Error("CLAUDE.md should contain worktree context")
	}
	if !strings.Contains(content, "team/dev") {
		t.Error("CLAUDE.md should contain branch name")
	}
	// NEW: should NOT contain old task references
	if strings.Contains(content, "tasks/pending") {
		t.Error("CLAUDE.md should not reference tasks/pending")
	}
	if strings.Contains(content, "tasks/done") {
		t.Error("CLAUDE.md should not reference tasks/done")
	}
}

func TestPromptMDTemplate(t *testing.T) {
	content := PromptMDContent("backend")
	if !strings.Contains(content, "# Role: backend") {
		t.Error("prompt template should contain role name")
	}
	if !strings.Contains(content, "Communication Protocol") {
		t.Error("prompt template should contain communication protocol")
	}
	// NEW: should contain Workflow section with OpenSpec
	if !strings.Contains(content, "## Workflow") {
		t.Error("prompt template should contain Workflow section")
	}
	if !strings.Contains(content, "/opsx:continue") {
		t.Error("prompt template should reference /opsx:continue")
	}
	if !strings.Contains(content, "/opsx:apply") {
		t.Error("prompt template should reference /opsx:apply")
	}
	// Should NOT contain old task references
	if strings.Contains(content, "tasks/pending") {
		t.Error("prompt template should not reference tasks/pending")
	}
}
