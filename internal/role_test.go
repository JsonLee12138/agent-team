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
	roleDir := filepath.Join(dir, ".agents", "teams", "backend")
	os.MkdirAll(roleDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# backend\n"), 0644)

	roles = ListAvailableRoles(dir)
	if len(roles) != 1 || roles[0] != "backend" {
		t.Errorf("ListAvailableRoles = %v, want [backend]", roles)
	}
}

func TestListWorkers(t *testing.T) {
	dir := t.TempDir()
	wtBase := ".worktrees"

	// empty → no workers
	workers := ListWorkers(dir, wtBase)
	if len(workers) != 0 {
		t.Errorf("ListWorkers(empty) = %v, want empty", workers)
	}

	// create a worktree dir with worker.yaml
	wtDir := filepath.Join(dir, wtBase, "backend-001")
	os.MkdirAll(wtDir, 0755)
	cfg := &WorkerConfig{WorkerID: "backend-001", Role: "backend", Provider: "claude"}
	cfg.Save(WorkerYAMLPath(wtDir))

	workers = ListWorkers(dir, wtBase)
	if len(workers) != 1 || workers[0].WorkerID != "backend-001" {
		t.Errorf("ListWorkers = %v, want [backend-001]", workers)
	}
	if workers[0].Role != "backend" {
		t.Errorf("ListWorkers[0].Role = %q, want backend", workers[0].Role)
	}
}

func TestNextWorkerID(t *testing.T) {
	dir := t.TempDir()
	wtBase := ".worktrees"

	// no worktrees → 001
	got := NextWorkerID(dir, wtBase, "frontend-dev")
	if got != "frontend-dev-001" {
		t.Errorf("NextWorkerID(empty) = %q, want frontend-dev-001", got)
	}

	// create worktree 001
	os.MkdirAll(filepath.Join(dir, wtBase, "frontend-dev-001"), 0755)
	got = NextWorkerID(dir, wtBase, "frontend-dev")
	if got != "frontend-dev-002" {
		t.Errorf("NextWorkerID(001 exists) = %q, want frontend-dev-002", got)
	}

	// create worktree 005 (gap)
	os.MkdirAll(filepath.Join(dir, wtBase, "frontend-dev-005"), 0755)
	got = NextWorkerID(dir, wtBase, "frontend-dev")
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
	for _, expected := range []string{".gitignore", ".claude/", ".codex/", ".tasks/", "worker.yaml"} {
		if !strings.Contains(content, expected) {
			t.Errorf(".gitignore should contain %q", expected)
		}
	}
}

func TestWorkerYAMLPath(t *testing.T) {
	got := WorkerYAMLPath("/tmp/wt/dev-001")
	want := "/tmp/wt/dev-001/worker.yaml"
	if got != want {
		t.Errorf("WorkerYAMLPath = %q, want %q", got, want)
	}
}

func TestBuildLaunchCmd(t *testing.T) {
	tests := []struct {
		provider, model, want string
	}{
		{"claude", "", "claude --dangerously-skip-permissions"},
		{"codex", "gpt-5", "codex --dangerously-bypass-approvals-and-sandbox --model gpt-5"},
		{"opencode", "", "opencode"},
		{"gemini", "", "gemini --approval-mode yolo"},
		{"gemini", "gemini-2.5-pro", "gemini --approval-mode yolo --model gemini-2.5-pro"},
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
		initial := "<!-- OTHER:START -->\nother content\n<!-- OTHER:END -->\n"
		os.WriteFile(fp, []byte(initial), 0644)

		err := InjectSection(fp, "AGENT_TEAM", "role prompt")
		if err != nil {
			t.Fatalf("InjectSection: %v", err)
		}

		data, _ := os.ReadFile(fp)
		content := string(data)
		if !strings.Contains(content, "other content") {
			t.Error("OTHER content should be preserved")
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

	// Create role system.md in .agents/teams/dev/
	roleDir := filepath.Join(dir, ".agents", "teams", "dev")
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
	if !strings.Contains(content, "Reply to main controller (used by workers)") {
		t.Error("CLAUDE.md should require reply-main protocol")
	}
	if !strings.Contains(content, "Task completed: <summary>") {
		t.Error("CLAUDE.md should contain completion reply example")
	}
	if !strings.Contains(content, "After the archive attempt (success or failure)") {
		t.Error("CLAUDE.md should require notification after archive attempt")
	}
	if !strings.Contains(content, "agent-team task archive") {
		t.Error("CLAUDE.md should contain task archive command")
	}
	if !strings.Contains(content, "Need decision: <problem or options>") {
		t.Error("CLAUDE.md should contain blocker/options reply example")
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

func TestResolveAgentsDir(t *testing.T) {
	t.Run("returns .agents/ when exists", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agents"), 0755)
		got := ResolveAgentsDir(dir)
		if got != filepath.Join(dir, ".agents") {
			t.Errorf("ResolveAgentsDir = %q, want .agents/", got)
		}
	})

	t.Run("falls back to agents/ when .agents/ missing", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "agents"), 0755)
		got := ResolveAgentsDir(dir)
		if got != filepath.Join(dir, "agents") {
			t.Errorf("ResolveAgentsDir = %q, want agents/", got)
		}
	})

	t.Run("prefers .agents/ over agents/ when both exist", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agents"), 0755)
		os.MkdirAll(filepath.Join(dir, "agents"), 0755)
		got := ResolveAgentsDir(dir)
		if got != filepath.Join(dir, ".agents") {
			t.Errorf("ResolveAgentsDir = %q, want .agents/ (priority)", got)
		}
	})

	t.Run("returns .agents/ when neither exists", func(t *testing.T) {
		dir := t.TempDir()
		got := ResolveAgentsDir(dir)
		if got != filepath.Join(dir, ".agents") {
			t.Errorf("ResolveAgentsDir = %q, want .agents/ (default)", got)
		}
	})
}

func TestRolePathFunctions(t *testing.T) {
	// Use a temp dir with .agents/ so ResolveAgentsDir returns .agents/
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".agents"), 0755)

	if got := RoleDir(root, "frontend-dev"); got != filepath.Join(root, ".agents", "teams", "frontend-dev") {
		t.Errorf("RoleDir = %q", got)
	}
}

func TestSplitRoleKeywords(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"frontend-dev", []string{"frontend", "dev"}},
		{"backend", []string{"backend"}},
		{"full-stack-engineer", []string{"full", "stack", "engineer"}},
		{"a-b-c", []string{"a", "b", "c"}},
	}
	for _, tt := range tests {
		got := splitRoleKeywords(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("splitRoleKeywords(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitRoleKeywords(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestResolveRole_ProjectFirst(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Create role in both project and global
	projectRole := filepath.Join(root, ".agents", "teams", "my-role")
	os.MkdirAll(projectRole, 0755)
	os.WriteFile(filepath.Join(projectRole, "SKILL.md"), []byte("# skill\n"), 0644)

	globalRole := filepath.Join(home, ".agents", "roles", "my-role")
	os.MkdirAll(globalRole, 0755)
	os.WriteFile(filepath.Join(globalRole, "SKILL.md"), []byte("# skill\n"), 0644)

	match, err := ResolveRole(root, "my-role")
	if err != nil {
		t.Fatalf("ResolveRole: %v", err)
	}
	if match.Scope != "project" {
		t.Errorf("Scope = %q, want project", match.Scope)
	}
	if match.Path != projectRole {
		t.Errorf("Path = %q, want %q", match.Path, projectRole)
	}
}

func TestResolveRole_FallbackGlobal(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)

	os.MkdirAll(filepath.Join(root, ".agents"), 0755)

	// Only global role exists
	globalRole := filepath.Join(home, ".agents", "roles", "my-role")
	os.MkdirAll(globalRole, 0755)
	os.WriteFile(filepath.Join(globalRole, "system.md"), []byte("# system\n"), 0644)

	match, err := ResolveRole(root, "my-role")
	if err != nil {
		t.Fatalf("ResolveRole: %v", err)
	}
	if match.Scope != "global" {
		t.Errorf("Scope = %q, want global", match.Scope)
	}
	if match.Path != globalRole {
		t.Errorf("Path = %q, want %q", match.Path, globalRole)
	}
}

func TestResolveRole_NotFound(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)

	os.MkdirAll(filepath.Join(root, ".agents"), 0755)

	_, err := ResolveRole(root, "nonexistent-role")
	if err == nil {
		t.Fatal("expected error for nonexistent role")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want 'not found' in message", err.Error())
	}
}

func TestSearchGlobalRoles_ExactMatch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Create global role
	rolePath := filepath.Join(home, ".agents", "roles", "frontend-dev")
	os.MkdirAll(rolePath, 0755)
	os.WriteFile(filepath.Join(rolePath, "SKILL.md"), []byte("# skill\n"), 0644)

	matches, err := SearchGlobalRoles("frontend-dev")
	if err != nil {
		t.Fatalf("SearchGlobalRoles: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	if matches[0].MatchType != "exact" {
		t.Errorf("MatchType = %q, want exact", matches[0].MatchType)
	}
	if matches[0].RoleName != "frontend-dev" {
		t.Errorf("RoleName = %q, want frontend-dev", matches[0].RoleName)
	}
}

func TestSearchGlobalRoles_KeywordMatch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Create global role with description containing keyword
	rolePath := filepath.Join(home, ".agents", "roles", "ui-engineer")
	refDir := filepath.Join(rolePath, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(rolePath, "SKILL.md"), []byte("# skill\n"), 0644)
	os.WriteFile(filepath.Join(refDir, "role.yaml"), []byte("description: \"Frontend UI development\"\n"), 0644)

	matches, err := SearchGlobalRoles("frontend-dev")
	if err != nil {
		t.Fatalf("SearchGlobalRoles: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	if matches[0].MatchType != "keyword" {
		t.Errorf("MatchType = %q, want keyword", matches[0].MatchType)
	}
	if matches[0].Description != "Frontend UI development" {
		t.Errorf("Description = %q, want 'Frontend UI development'", matches[0].Description)
	}
}

func TestListGlobalRoles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Create two global roles
	for _, name := range []string{"backend-dev", "frontend-dev"} {
		rolePath := filepath.Join(home, ".agents", "roles", name)
		os.MkdirAll(rolePath, 0755)
		os.WriteFile(filepath.Join(rolePath, "SKILL.md"), []byte("# skill\n"), 0644)
	}

	roles, err := ListGlobalRoles()
	if err != nil {
		t.Fatalf("ListGlobalRoles: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("got %d roles, want 2", len(roles))
	}
	if roles[0].RoleName != "backend-dev" {
		t.Errorf("roles[0].RoleName = %q, want backend-dev", roles[0].RoleName)
	}
	if roles[1].RoleName != "frontend-dev" {
		t.Errorf("roles[1].RoleName = %q, want frontend-dev", roles[1].RoleName)
	}
}

func TestListGlobalRoles_DirNotExists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	roles, err := ListGlobalRoles()
	if err != nil {
		t.Fatalf("ListGlobalRoles: %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("got %d roles, want 0", len(roles))
	}
}
