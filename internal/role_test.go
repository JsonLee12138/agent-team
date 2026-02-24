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
		// Should have exactly one start marker
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

func TestInjectRolePrompt(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, ".worktrees", "dev")
	teamsDir := filepath.Join(wtPath, "agents", "teams", "dev")
	os.MkdirAll(teamsDir, 0755)

	promptContent := "# Role: dev\n\nA developer role.\n"
	os.WriteFile(filepath.Join(teamsDir, "prompt.md"), []byte(promptContent), 0644)

	err := InjectRolePrompt(wtPath, "dev", dir)
	if err != nil {
		t.Fatalf("InjectRolePrompt: %v", err)
	}

	// Check CLAUDE.md
	data, _ := os.ReadFile(filepath.Join(wtPath, "CLAUDE.md"))
	content := string(data)

	if !strings.Contains(content, "<!-- AGENT_TEAM:START -->") {
		t.Error("CLAUDE.md should contain AGENT_TEAM start marker")
	}
	if !strings.Contains(content, "<!-- AGENT_TEAM:END -->") {
		t.Error("CLAUDE.md should contain AGENT_TEAM end marker")
	}
	if !strings.Contains(content, "# Role: dev") {
		t.Error("CLAUDE.md should contain prompt.md content")
	}
	if !strings.Contains(content, "Development Environment") {
		t.Error("CLAUDE.md should contain worktree context")
	}
	if !strings.Contains(content, "team/dev") {
		t.Error("CLAUDE.md should contain branch name")
	}

	// Check AGENTS.md
	agentsData, _ := os.ReadFile(filepath.Join(wtPath, "AGENTS.md"))
	agentsContent := string(agentsData)

	if !strings.Contains(agentsContent, "<!-- AGENT_TEAM:START -->") {
		t.Error("AGENTS.md should contain AGENT_TEAM start marker")
	}
	if !strings.Contains(agentsContent, "# Role: dev") {
		t.Error("AGENTS.md should contain prompt.md content")
	}
}

func TestInjectRolePromptPreservesOpenSpec(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, ".worktrees", "dev")
	teamsDir := filepath.Join(wtPath, "agents", "teams", "dev")
	os.MkdirAll(teamsDir, 0755)

	promptContent := "# Role: dev\n\nA developer role.\n"
	os.WriteFile(filepath.Join(teamsDir, "prompt.md"), []byte(promptContent), 0644)

	// Simulate OpenSpec having already written to CLAUDE.md
	openspecContent := "<!-- OPENSPEC:START -->\nOpenSpec config here\n<!-- OPENSPEC:END -->\n"
	os.WriteFile(filepath.Join(wtPath, "CLAUDE.md"), []byte(openspecContent), 0644)
	os.WriteFile(filepath.Join(wtPath, "AGENTS.md"), []byte(openspecContent), 0644)

	err := InjectRolePrompt(wtPath, "dev", dir)
	if err != nil {
		t.Fatalf("InjectRolePrompt: %v", err)
	}

	// Check CLAUDE.md preserves OpenSpec
	data, _ := os.ReadFile(filepath.Join(wtPath, "CLAUDE.md"))
	content := string(data)

	if !strings.Contains(content, "OpenSpec config here") {
		t.Error("CLAUDE.md should preserve OPENSPEC content")
	}
	if !strings.Contains(content, "<!-- AGENT_TEAM:START -->") {
		t.Error("CLAUDE.md should contain AGENT_TEAM section")
	}

	// Check AGENTS.md preserves OpenSpec
	agentsData, _ := os.ReadFile(filepath.Join(wtPath, "AGENTS.md"))
	agentsContent := string(agentsData)

	if !strings.Contains(agentsContent, "OpenSpec config here") {
		t.Error("AGENTS.md should preserve OPENSPEC content")
	}
	if !strings.Contains(agentsContent, "<!-- AGENT_TEAM:START -->") {
		t.Error("AGENTS.md should contain AGENT_TEAM section")
	}
}

func TestInjectRolePromptUpdate(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, ".worktrees", "dev")
	teamsDir := filepath.Join(wtPath, "agents", "teams", "dev")
	os.MkdirAll(teamsDir, 0755)

	// First injection
	os.WriteFile(filepath.Join(teamsDir, "prompt.md"), []byte("# Role: dev v1\n"), 0644)
	InjectRolePrompt(wtPath, "dev", dir)

	// Update prompt and re-inject
	os.WriteFile(filepath.Join(teamsDir, "prompt.md"), []byte("# Role: dev v2\n"), 0644)
	err := InjectRolePrompt(wtPath, "dev", dir)
	if err != nil {
		t.Fatalf("InjectRolePrompt (update): %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(wtPath, "CLAUDE.md"))
	content := string(data)

	if strings.Contains(content, "dev v1") {
		t.Error("old prompt content should be replaced")
	}
	if !strings.Contains(content, "dev v2") {
		t.Error("new prompt content should be present")
	}
	if strings.Count(content, "<!-- AGENT_TEAM:START -->") != 1 {
		t.Error("should have exactly one AGENT_TEAM section")
	}
}

func TestInjectRolePromptNoPromptFile(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, ".worktrees", "dev")
	os.MkdirAll(wtPath, 0755)

	// No prompt.md — should be a no-op
	err := InjectRolePrompt(wtPath, "dev", dir)
	if err != nil {
		t.Fatalf("InjectRolePrompt should not error without prompt.md: %v", err)
	}

	// CLAUDE.md should not be created
	if _, err := os.Stat(filepath.Join(wtPath, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Error("CLAUDE.md should not be created without prompt.md")
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
