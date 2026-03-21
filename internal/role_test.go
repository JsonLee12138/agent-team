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
	roleDir := filepath.Join(dir, ".agent-team", "teams", "backend")
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
	os.MkdirAll(filepath.Join(dir, ".agent-team"), 0755)

	// empty → no workers
	workers := ListWorkers(dir, wtBase)
	if len(workers) != 0 {
		t.Errorf("ListWorkers(empty) = %v, want empty", workers)
	}

	central := &WorkerConfig{WorkerID: "backend-001", Role: "backend", Provider: "claude"}
	if err := central.Save(WorkerConfigPath(dir, "backend-001")); err != nil {
		t.Fatalf("save centralized config: %v", err)
	}

	legacyDir := filepath.Join(dir, wtBase, "backend-002")
	os.MkdirAll(legacyDir, 0755)
	legacy := &WorkerConfig{WorkerID: "backend-002", Role: "backend", Provider: "claude"}
	if err := legacy.Save(WorkerYAMLPath(legacyDir)); err != nil {
		t.Fatalf("save legacy config: %v", err)
	}

	duplicateDir := filepath.Join(dir, wtBase, "backend-001")
	os.MkdirAll(duplicateDir, 0755)
	duplicate := &WorkerConfig{WorkerID: "backend-001", Role: "legacy-duplicate", Provider: "claude"}
	if err := duplicate.Save(WorkerYAMLPath(duplicateDir)); err != nil {
		t.Fatalf("save duplicate legacy config: %v", err)
	}

	workers = ListWorkers(dir, wtBase)
	if len(workers) != 2 {
		t.Fatalf("ListWorkers = %v, want 2 workers", workers)
	}
	if workers[0].WorkerID != "backend-001" || workers[0].Role != "legacy-duplicate" {
		t.Fatalf("workers[0] = %+v", workers[0])
	}
	if workers[1].WorkerID != "backend-002" || workers[1].Role != "backend" {
		t.Fatalf("workers[1] = %+v", workers[1])
	}
}

func TestNextWorkerID(t *testing.T) {
	dir := t.TempDir()
	wtBase := ".worktrees"
	os.MkdirAll(filepath.Join(dir, ".agent-team"), 0755)

	// no workers → 001
	got := NextWorkerID(dir, wtBase, "frontend-dev")
	if got != "frontend-dev-001" {
		t.Errorf("NextWorkerID(empty) = %q, want frontend-dev-001", got)
	}

	// centralized 001
	if err := os.MkdirAll(WorkerConfigDir(dir, "frontend-dev-001"), 0755); err != nil {
		t.Fatalf("mkdir centralized worker: %v", err)
	}
	got = NextWorkerID(dir, wtBase, "frontend-dev")
	if got != "frontend-dev-002" {
		t.Errorf("NextWorkerID(centralized 001 exists) = %q, want frontend-dev-002", got)
	}

	// legacy 005
	os.MkdirAll(filepath.Join(dir, wtBase, "frontend-dev-005"), 0755)
	got = NextWorkerID(dir, wtBase, "frontend-dev")
	if got != "frontend-dev-006" {
		t.Errorf("NextWorkerID(legacy 005 exists) = %q, want frontend-dev-006", got)
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
	for _, expected := range []string{".gitignore", ".claude/", ".codex/", ".tasks/", "worker.yaml", "CLAUDE.md", "GEMINI.md", "AGENTS.md"} {
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

func TestInjectRolePromptLegacy(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, ".worktrees", "dev-001")
	os.MkdirAll(wtPath, 0755)

	// Create role system.md in .agent-team/teams/dev/ (NO .agent-team/rules/ → legacy mode)
	roleDir := filepath.Join(dir, ".agent-team", "teams", "dev")
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
	if !strings.Contains(content, "Completion Protocol") {
		t.Error("CLAUDE.md should contain task completion protocol (legacy)")
	}
	if !strings.Contains(content, "Use `agent-team reply-main` for worker-to-main communication") {
		t.Error("CLAUDE.md should require reply-main protocol (legacy)")
	}
	if !strings.Contains(content, "Task completed: <summary>") {
		t.Error("CLAUDE.md should contain completion reply example (legacy)")
	}
	if !strings.Contains(content, "agent-team reply-main \"Task completed: <summary>\"") {
		t.Error("CLAUDE.md should contain reply-main completion command (legacy)")
	}
	if !strings.Contains(content, "Need decision: <problem or options>") {
		t.Error("CLAUDE.md should contain blocker/options reply example (legacy)")
	}
}

func TestInjectRolePromptSlim(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, ".worktrees", "dev-001")
	os.MkdirAll(wtPath, 0755)

	// Create role with system.md and role.yaml
	roleDir := filepath.Join(dir, ".agent-team", "teams", "dev")
	refDir := filepath.Join(roleDir, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "system.md"), []byte("# System Prompt: dev\n\nA developer role.\n"), 0644)
	os.WriteFile(filepath.Join(refDir, "role.yaml"), []byte("name: dev\ndescription: \"Full-stack developer\"\nskills:\n  - name: \"systematic-debugging\"\n    description: \"Debugging workflow\"\n"), 0644)

	// Create SKILL.md with frontmatter trigger
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("---\nname: dev\ndescription: >\n  Full-stack dev skill.\n  Use when the user asks for dev work.\n---\n# dev\n"), 0644)

	// Create .agent-team/rules/ directory with index.md → triggers slim mode
	rulesDir := filepath.Join(dir, ".agent-team", "rules")
	os.MkdirAll(rulesDir, 0755)
	os.WriteFile(filepath.Join(rulesDir, "index.md"), []byte("- `debugging.md`: bug, flaky test\n- `worktree.md`: task start, verify\n"), 0644)

	err := InjectRolePrompt(wtPath, "dev-001", "dev", dir)
	if err != nil {
		t.Fatalf("InjectRolePrompt: %v", err)
	}

	// Check all three provider files get the same content
	for _, name := range []string{"CLAUDE.md", "AGENTS.md", "GEMINI.md"} {
		data, err := os.ReadFile(filepath.Join(wtPath, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		content := string(data)

		if !strings.Contains(content, "<!-- AGENT_TEAM:START -->") {
			t.Errorf("%s should contain AGENT_TEAM start marker", name)
		}
		// Slim mode: minimal identity (generated, not full system.md)
		if !strings.Contains(content, "You are the dev role") {
			t.Errorf("%s should contain role identity line", name)
		}
		if !strings.Contains(content, "Full-stack developer") {
			t.Errorf("%s should contain description from role.yaml", name)
		}
		// Slim mode: Development Environment present but Git Rules / Completion Protocol NOT inlined
		if !strings.Contains(content, "Development Environment") {
			t.Errorf("%s should contain worktree context", name)
		}
		if strings.Contains(content, "### Git Rules") {
			t.Errorf("%s should NOT inline Git Rules in slim mode", name)
		}
		if strings.Contains(content, "### Completion Protocol") {
			t.Errorf("%s should NOT inline Completion Protocol in slim mode", name)
		}
		// Rules index section present
		if !strings.Contains(content, "Rules Reference") {
			t.Errorf("%s should contain Rules Reference section", name)
		}
		if !strings.Contains(content, "debugging.md") {
			t.Errorf("%s should contain rules index content", name)
		}
		// Skill index section present
		if !strings.Contains(content, "Skill Index") {
			t.Errorf("%s should contain Skill Index section", name)
		}
		if !strings.Contains(content, "**dev**") {
			t.Errorf("%s should list role skill in index", name)
		}
	}
}

func TestInjectRolePromptSlimShorterThanLegacy(t *testing.T) {
	dir := t.TempDir()

	// Setup role
	roleDir := filepath.Join(dir, ".agent-team", "teams", "dev")
	refDir := filepath.Join(roleDir, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "system.md"), []byte("# System Prompt: dev\n\nA developer role.\n"), 0644)
	os.WriteFile(filepath.Join(refDir, "role.yaml"), []byte("name: dev\ndescription: \"Full-stack developer\"\n"), 0644)

	// Build legacy content (no rules dir)
	legacyWT := filepath.Join(dir, ".worktrees", "legacy-001")
	os.MkdirAll(legacyWT, 0755)
	err := InjectRolePrompt(legacyWT, "legacy-001", "dev", dir)
	if err != nil {
		t.Fatalf("legacy inject: %v", err)
	}
	legacyData, _ := os.ReadFile(filepath.Join(legacyWT, "CLAUDE.md"))
	legacyLen := len(legacyData)

	// Now create rules dir and build slim content
	rulesDir := filepath.Join(dir, ".agent-team", "rules")
	os.MkdirAll(rulesDir, 0755)
	os.WriteFile(filepath.Join(rulesDir, "index.md"), []byte("- `debugging.md`: bug\n"), 0644)

	slimWT := filepath.Join(dir, ".worktrees", "slim-001")
	os.MkdirAll(slimWT, 0755)
	err = InjectRolePrompt(slimWT, "slim-001", "dev", dir)
	if err != nil {
		t.Fatalf("slim inject: %v", err)
	}
	slimData, _ := os.ReadFile(filepath.Join(slimWT, "CLAUDE.md"))
	slimLen := len(slimData)

	if slimLen >= legacyLen {
		t.Errorf("slim mode (%d bytes) should be shorter than legacy mode (%d bytes)", slimLen, legacyLen)
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
		os.MkdirAll(filepath.Join(dir, ".agent-team"), 0755)
		got := ResolveAgentsDir(dir)
		if got != filepath.Join(dir, ".agent-team") {
			t.Errorf("ResolveAgentsDir = %q, want .agents/", got)
		}
	})

	t.Run("ignores legacy agents/ and still returns .agent-team", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "agents"), 0755)
		got := ResolveAgentsDir(dir)
		if got != filepath.Join(dir, ".agent-team") {
			t.Errorf("ResolveAgentsDir = %q, want .agent-team/", got)
		}
	})

	t.Run("prefers .agents/ over agents/ when both exist", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agent-team"), 0755)
		os.MkdirAll(filepath.Join(dir, "agents"), 0755)
		got := ResolveAgentsDir(dir)
		if got != filepath.Join(dir, ".agent-team") {
			t.Errorf("ResolveAgentsDir = %q, want .agents/ (priority)", got)
		}
	})

	t.Run("returns .agents/ when neither exists", func(t *testing.T) {
		dir := t.TempDir()
		got := ResolveAgentsDir(dir)
		if got != filepath.Join(dir, ".agent-team") {
			t.Errorf("ResolveAgentsDir = %q, want .agents/ (default)", got)
		}
	})
}

func TestRolePathFunctions(t *testing.T) {
	// Use a temp dir with .agents/ so ResolveAgentsDir returns .agents/
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".agent-team"), 0755)

	if got := RoleDir(root, "frontend-dev"); got != filepath.Join(root, ".agent-team", "teams", "frontend-dev") {
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
	projectRole := filepath.Join(root, ".agent-team", "teams", "my-role")
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

	os.MkdirAll(filepath.Join(root, ".agent-team"), 0755)

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

	os.MkdirAll(filepath.Join(root, ".agent-team"), 0755)

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

func TestHasRulesDir(t *testing.T) {
	t.Run("returns true when .agent-team/rules/ exists", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agent-team", "rules"), 0755)
		if !HasRulesDir(dir) {
			t.Error("HasRulesDir should return true")
		}
	})

	t.Run("returns false when .agent-team/rules/ missing", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agent-team"), 0755)
		if HasRulesDir(dir) {
			t.Error("HasRulesDir should return false")
		}
	})

	t.Run("returns false when rules is a file not dir", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agent-team"), 0755)
		os.WriteFile(filepath.Join(dir, ".agent-team", "rules"), []byte("not a dir"), 0644)
		if HasRulesDir(dir) {
			t.Error("HasRulesDir should return false when rules is a file")
		}
	})
}

func TestBuildRoleIdentity(t *testing.T) {
	dir := t.TempDir()
	roleDir := filepath.Join(dir, ".agent-team", "teams", "qa-tester")
	refDir := filepath.Join(roleDir, "references")
	os.MkdirAll(refDir, 0755)

	// With description in role.yaml
	os.WriteFile(filepath.Join(refDir, "role.yaml"), []byte("description: \"Write tests and verify quality\"\n"), 0644)

	identity := buildRoleIdentity("qa-tester", roleDir)
	if !strings.Contains(identity, "You are the qa-tester role") {
		t.Error("identity should contain role name")
	}
	if !strings.Contains(identity, "Write tests and verify quality") {
		t.Error("identity should contain description from role.yaml")
	}
}

func TestBuildRoleIdentityNoDescription(t *testing.T) {
	dir := t.TempDir()
	roleDir := filepath.Join(dir, ".agent-team", "teams", "minimal")
	os.MkdirAll(roleDir, 0755)
	// No references/role.yaml

	identity := buildRoleIdentity("minimal", roleDir)
	if !strings.Contains(identity, "You are the minimal role") {
		t.Error("identity should contain role name")
	}
	if strings.Contains(identity, "Primary objective") {
		t.Error("identity should NOT contain 'Primary objective' when no description")
	}
}

func TestBuildRulesIndexSection(t *testing.T) {
	t.Run("returns content when index.md exists", func(t *testing.T) {
		dir := t.TempDir()
		rulesDir := filepath.Join(dir, ".agent-team", "rules")
		os.MkdirAll(rulesDir, 0755)
		os.WriteFile(filepath.Join(rulesDir, "index.md"), []byte("- `debugging.md`: bugs\n- `worktree.md`: tasks\n"), 0644)

		section := buildRulesIndexSection(dir)
		if !strings.Contains(section, "Rules Reference") {
			t.Error("should contain 'Rules Reference' header")
		}
		if !strings.Contains(section, "debugging.md") {
			t.Error("should contain rule file reference")
		}
		if !strings.Contains(section, "worktree.md") {
			t.Error("should contain task-protocol reference")
		}
	})

	t.Run("returns empty when index.md missing", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agent-team", "rules"), 0755)
		// No index.md

		section := buildRulesIndexSection(dir)
		if section != "" {
			t.Errorf("should return empty string, got %q", section)
		}
	})

	t.Run("returns empty when rules dir missing", func(t *testing.T) {
		dir := t.TempDir()
		section := buildRulesIndexSection(dir)
		if section != "" {
			t.Errorf("should return empty string, got %q", section)
		}
	})
}

func TestExtractSkillTrigger(t *testing.T) {
	t.Run("extracts description from frontmatter", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: test\ndescription: >\n  My test skill.\n  Use for testing.\n---\n# test\n"), 0644)

		trigger := extractSkillTrigger(dir)
		if trigger == "" {
			t.Fatal("trigger should not be empty")
		}
		if !strings.Contains(trigger, "My test skill") {
			t.Errorf("trigger = %q, should contain 'My test skill'", trigger)
		}
	})

	t.Run("returns empty without frontmatter", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# test\nNo frontmatter here.\n"), 0644)

		trigger := extractSkillTrigger(dir)
		if trigger != "" {
			t.Errorf("trigger should be empty, got %q", trigger)
		}
	})

	t.Run("returns empty when SKILL.md missing", func(t *testing.T) {
		dir := t.TempDir()
		trigger := extractSkillTrigger(dir)
		if trigger != "" {
			t.Errorf("trigger should be empty, got %q", trigger)
		}
	})
}

func TestInjectRolePromptSlimNoGitRulesOrTaskProtocol(t *testing.T) {
	// Explicitly verify slim mode excludes Git Rules and Completion Protocol content
	dir := t.TempDir()
	wtPath := filepath.Join(dir, ".worktrees", "test-001")
	os.MkdirAll(wtPath, 0755)

	roleDir := filepath.Join(dir, ".agent-team", "teams", "test")
	refDir := filepath.Join(roleDir, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "system.md"), []byte("# System Prompt: test\n"), 0644)
	os.WriteFile(filepath.Join(refDir, "role.yaml"), []byte("name: test\ndescription: Test role\n"), 0644)

	// Create rules dir → triggers slim mode
	rulesDir := filepath.Join(dir, ".agent-team", "rules")
	os.MkdirAll(rulesDir, 0755)
	os.WriteFile(filepath.Join(rulesDir, "index.md"), []byte("- `worktree.md`: tasks\n"), 0644)

	if err := InjectRolePrompt(wtPath, "test-001", "test", dir); err != nil {
		t.Fatalf("InjectRolePrompt: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(wtPath, "CLAUDE.md"))
	content := string(data)

	// Should NOT contain legacy inline sections
	forbiddenPhrases := []string{
		"### Git Rules",
		"### Completion Protocol",
		"git add -A",
		"agent-team reply-main \"Task completed: <summary>\"",
		"Reply to main controller",
		"Need decision: <problem or options>",
	}
	for _, phrase := range forbiddenPhrases {
		if strings.Contains(content, phrase) {
			t.Errorf("slim mode should NOT contain %q", phrase)
		}
	}

	// Should contain slim-mode sections
	requiredPhrases := []string{
		"You are the test role",
		"Development Environment",
		"Rules Reference",
		"worktree.md",
	}
	for _, phrase := range requiredPhrases {
		if !strings.Contains(content, phrase) {
			t.Errorf("slim mode should contain %q", phrase)
		}
	}
}

func TestInjectRolePromptLegacyContainsFullContent(t *testing.T) {
	// Verify legacy mode (no rules dir) contains full inline Git Rules and Completion Protocol
	dir := t.TempDir()
	wtPath := filepath.Join(dir, ".worktrees", "legacy-001")
	os.MkdirAll(wtPath, 0755)

	roleDir := filepath.Join(dir, ".agent-team", "teams", "dev")
	os.MkdirAll(roleDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "system.md"), []byte("# System Prompt: dev\nA developer.\n"), 0644)

	// No .agent-team/rules/ → legacy mode
	if err := InjectRolePrompt(wtPath, "legacy-001", "dev", dir); err != nil {
		t.Fatalf("InjectRolePrompt: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(wtPath, "CLAUDE.md"))
	content := string(data)

	requiredPhrases := []string{
		"# System Prompt: dev",
		"### Git Rules",
		"### Completion Protocol",
		"agent-team reply-main \"Task completed: <summary>\"",
		"Use `agent-team reply-main` for worker-to-main communication",
		"Need decision: <problem or options>",
	}
	for _, phrase := range requiredPhrases {
		if !strings.Contains(content, phrase) {
			t.Errorf("legacy mode should contain %q", phrase)
		}
	}

	// Legacy should NOT contain Rules Reference section
	if strings.Contains(content, "Rules Reference") {
		t.Error("legacy mode should NOT contain 'Rules Reference'")
	}
}

func TestInjectRolePromptWithPathExplicit(t *testing.T) {
	// Test InjectRolePromptWithPath using an explicit role path
	dir := t.TempDir()
	wtPath := filepath.Join(dir, ".worktrees", "wp-001")
	os.MkdirAll(wtPath, 0755)

	// Place role outside the standard .agent-team/teams/ location
	customRolePath := filepath.Join(dir, "custom-roles", "my-role")
	os.MkdirAll(customRolePath, 0755)
	os.WriteFile(filepath.Join(customRolePath, "system.md"), []byte("# System Prompt: my-role\nCustom location.\n"), 0644)

	err := InjectRolePromptWithPath(wtPath, "wp-001", "my-role", customRolePath, dir)
	if err != nil {
		t.Fatalf("InjectRolePromptWithPath: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(wtPath, "CLAUDE.md"))
	content := string(data)
	if !strings.Contains(content, "# System Prompt: my-role") {
		t.Error("should inject content from explicit role path")
	}
	if !strings.Contains(content, "Custom location") {
		t.Error("should include system.md content from explicit path")
	}
}

func TestInjectRolePromptAllThreeProviderFiles(t *testing.T) {
	// Verify all three provider files are created with consistent content
	dir := t.TempDir()
	wtPath := filepath.Join(dir, ".worktrees", "multi-001")
	os.MkdirAll(wtPath, 0755)

	roleDir := filepath.Join(dir, ".agent-team", "teams", "multi")
	os.MkdirAll(roleDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "system.md"), []byte("# System Prompt: multi\n"), 0644)

	if err := InjectRolePrompt(wtPath, "multi-001", "multi", dir); err != nil {
		t.Fatalf("InjectRolePrompt: %v", err)
	}

	var contents [3]string
	for i, name := range []string{"CLAUDE.md", "AGENTS.md", "GEMINI.md"} {
		data, err := os.ReadFile(filepath.Join(wtPath, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		contents[i] = string(data)
		if !strings.Contains(contents[i], "AGENT_TEAM:START") {
			t.Errorf("%s missing AGENT_TEAM markers", name)
		}
	}

	// All three files should have identical content
	if contents[0] != contents[1] {
		t.Error("CLAUDE.md and AGENTS.md should have identical content")
	}
	if contents[0] != contents[2] {
		t.Error("CLAUDE.md and GEMINI.md should have identical content")
	}
}
