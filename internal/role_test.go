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

func TestListAvailableRoles(t *testing.T) {
	dir := t.TempDir()

	// empty → no roles
	roles := ListAvailableRoles(dir)
	if len(roles) != 0 {
		t.Errorf("ListAvailableRoles(empty) = %v, want empty", roles)
	}

	// create a role with SKILL.md
	roleDir := filepath.Join(dir, "agents", "teams", "backend")
	os.MkdirAll(roleDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# backend\n"), 0644)

	roles = ListAvailableRoles(dir)
	if len(roles) != 1 || roles[0] != "backend" {
		t.Errorf("ListAvailableRoles = %v, want [backend]", roles)
	}
}

func TestListWorkers(t *testing.T) {
	dir := t.TempDir()

	// empty → no workers
	workers := ListWorkers(dir)
	if len(workers) != 0 {
		t.Errorf("ListWorkers(empty) = %v, want empty", workers)
	}

	// create a worker
	workerDir := filepath.Join(dir, "agents", "workers", "backend-001")
	os.MkdirAll(workerDir, 0755)
	cfg := &WorkerConfig{WorkerID: "backend-001", Role: "backend"}
	cfg.Save(filepath.Join(workerDir, "config.yaml"))

	workers = ListWorkers(dir)
	if len(workers) != 1 || workers[0].WorkerID != "backend-001" {
		t.Errorf("ListWorkers = %v, want [backend-001]", workers)
	}
	if workers[0].Role != "backend" {
		t.Errorf("ListWorkers[0].Role = %q, want backend", workers[0].Role)
	}
}

func TestNextWorkerID(t *testing.T) {
	dir := t.TempDir()

	// no workers → 001
	got := NextWorkerID(dir, "frontend-dev")
	if got != "frontend-dev-001" {
		t.Errorf("NextWorkerID(empty) = %q, want frontend-dev-001", got)
	}

	// create worker 001
	os.MkdirAll(filepath.Join(dir, "agents", "workers", "frontend-dev-001"), 0755)
	got = NextWorkerID(dir, "frontend-dev")
	if got != "frontend-dev-002" {
		t.Errorf("NextWorkerID(001 exists) = %q, want frontend-dev-002", got)
	}

	// create worker 005 (gap)
	os.MkdirAll(filepath.Join(dir, "agents", "workers", "frontend-dev-005"), 0755)
	got = NextWorkerID(dir, "frontend-dev")
	if got != "frontend-dev-006" {
		t.Errorf("NextWorkerID(005 exists) = %q, want frontend-dev-006", got)
	}
}

func TestWriteWorktreeGitignore(t *testing.T) {
	dir := t.TempDir()
	if err := WriteWorktreeGitignore(dir); err != nil {
		t.Fatalf("WriteWorktreeGitignore: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	content := string(data)
	for _, expected := range []string{".gitignore", ".claude/", ".codex/", "openspec/"} {
		if !strings.Contains(content, expected) {
			t.Errorf(".gitignore should contain %q", expected)
		}
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

func TestInjectSection(t *testing.T) {
	t.Run("creates new file if not exists", func(t *testing.T) {
		dir := t.TempDir()
		fp := filepath.Join(dir, "test.md")

		err := InjectSection(fp, "MY_TAG", "hello world")
		if err != nil {
			t.Fatalf("InjectSection: %v", err)
		}

		data, _ := os.ReadFile(fp)
		content := string(data)
		if !strings.Contains(content, "<!-- MY_TAG:START -->") {
			t.Error("should contain start marker")
		}
		if !strings.Contains(content, "hello world") {
			t.Error("should contain injected content")
		}
		if !strings.Contains(content, "<!-- MY_TAG:END -->") {
			t.Error("should contain end marker")
		}
	})

	t.Run("prepends to existing file without markers", func(t *testing.T) {
		dir := t.TempDir()
		fp := filepath.Join(dir, "test.md")
		os.WriteFile(fp, []byte("existing content\n"), 0644)

		err := InjectSection(fp, "MY_TAG", "injected")
		if err != nil {
			t.Fatalf("InjectSection: %v", err)
		}

		data, _ := os.ReadFile(fp)
		content := string(data)
		if !strings.HasPrefix(content, "<!-- MY_TAG:START -->") {
			t.Error("injected section should be prepended")
		}
		if !strings.Contains(content, "existing content") {
			t.Error("existing content should be preserved")
		}
	})

	t.Run("replaces existing section", func(t *testing.T) {
		dir := t.TempDir()
		fp := filepath.Join(dir, "test.md")
		initial := "<!-- MY_TAG:START -->\nold content\n<!-- MY_TAG:END -->\n\nother stuff\n"
		os.WriteFile(fp, []byte(initial), 0644)

		err := InjectSection(fp, "MY_TAG", "new content")
		if err != nil {
			t.Fatalf("InjectSection: %v", err)
		}

		data, _ := os.ReadFile(fp)
		content := string(data)
		if strings.Contains(content, "old content") {
			t.Error("old content should be replaced")
		}
		if !strings.Contains(content, "new content") {
			t.Error("new content should be present")
		}
		if !strings.Contains(content, "other stuff") {
			t.Error("other content should be preserved")
		}
		if strings.Count(content, "<!-- MY_TAG:START -->") != 1 {
			t.Error("should have exactly one start marker")
		}
	})

	t.Run("preserves other tags", func(t *testing.T) {
		dir := t.TempDir()
		fp := filepath.Join(dir, "test.md")
		initial := "<!-- OPENSPEC:START -->\nopenspec content\n<!-- OPENSPEC:END -->\n"
		os.WriteFile(fp, []byte(initial), 0644)

		err := InjectSection(fp, "AGENT_TEAM", "role prompt")
		if err != nil {
			t.Fatalf("InjectSection: %v", err)
		}

		data, _ := os.ReadFile(fp)
		content := string(data)
		if !strings.Contains(content, "openspec content") {
			t.Error("OPENSPEC content should be preserved")
		}
		if !strings.Contains(content, "role prompt") {
			t.Error("AGENT_TEAM content should be injected")
		}
	})
}

func TestInjectRolePromptV2(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, ".worktrees", "dev-001")
	os.MkdirAll(wtPath, 0755)

	// Create role system.md in agents/teams/dev/
	roleDir := filepath.Join(dir, "agents", "teams", "dev")
	os.MkdirAll(roleDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "system.md"), []byte("# System Prompt: dev\n\nA developer role.\n"), 0644)

	err := InjectRolePrompt(wtPath, "dev-001", "dev", dir)
	if err != nil {
		t.Fatalf("InjectRolePrompt: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(wtPath, "CLAUDE.md"))
	content := string(data)

	if !strings.Contains(content, "<!-- AGENT_TEAM:START -->") {
		t.Error("CLAUDE.md should contain AGENT_TEAM start marker")
	}
	if !strings.Contains(content, "# System Prompt: dev") {
		t.Error("CLAUDE.md should contain system.md content")
	}
	if !strings.Contains(content, "Development Environment") {
		t.Error("CLAUDE.md should contain worktree context")
	}
	if !strings.Contains(content, "team/dev-001") {
		t.Error("CLAUDE.md should contain worker branch name")
	}
	if !strings.Contains(content, "Task Completion Protocol") {
		t.Error("CLAUDE.md should contain task completion protocol")
	}
}

func TestInjectRolePromptNoSource(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, ".worktrees", "dev-001")
	os.MkdirAll(wtPath, 0755)

	// No system.md, no prompt.md — should be a no-op
	err := InjectRolePrompt(wtPath, "dev-001", "dev", dir)
	if err != nil {
		t.Fatalf("InjectRolePrompt should not error without source: %v", err)
	}

	if _, err := os.Stat(filepath.Join(wtPath, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Error("CLAUDE.md should not be created without source")
	}
}

func TestRolePathFunctions(t *testing.T) {
	root := "/project"

	if got := RoleDir(root, "frontend-dev"); got != "/project/agents/teams/frontend-dev" {
		t.Errorf("RoleDir = %q", got)
	}
	if got := WorkerDir(root, "frontend-dev-001"); got != "/project/agents/workers/frontend-dev-001" {
		t.Errorf("WorkerDir = %q", got)
	}
	if got := WorkerConfigPath(root, "frontend-dev-001"); got != "/project/agents/workers/frontend-dev-001/config.yaml" {
		t.Errorf("WorkerConfigPath = %q", got)
	}
}
