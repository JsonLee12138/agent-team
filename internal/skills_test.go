// internal/skills_test.go
package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSkillName(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"vite", "vite"},
		{"antfu/skills@vite", "vite"},
		{"jsonlee12138/prompts@eslint-config", "eslint-config"},
		{"design-patterns-principles", "design-patterns-principles"},
		{"org/repo@sub@deep", "deep"},
	}
	for _, tt := range tests {
		got := parseSkillName(tt.input)
		if got != tt.want {
			t.Errorf("parseSkillName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestReadRoleSkills(t *testing.T) {
	dir := t.TempDir()

	t.Run("missing file returns nil", func(t *testing.T) {
		skills, err := ReadRoleSkills(dir, "nonexistent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if skills != nil {
			t.Errorf("expected nil, got %v", skills)
		}
	})

	t.Run("reads skills list", func(t *testing.T) {
		roleDir := filepath.Join(dir, ".agents", "teams", "dev", "references")
		os.MkdirAll(roleDir, 0755)
		content := "name: dev\nskills:\n  - vite\n  - antfu/skills@vitest\n"
		os.WriteFile(filepath.Join(roleDir, "role.yaml"), []byte(content), 0644)

		skills, err := ReadRoleSkills(dir, "dev")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(skills) != 2 {
			t.Fatalf("expected 2 skills, got %d", len(skills))
		}
		if skills[0] != "vite" || skills[1] != "antfu/skills@vitest" {
			t.Errorf("skills = %v", skills)
		}
	})

	t.Run("empty skills returns empty", func(t *testing.T) {
		roleDir := filepath.Join(dir, ".agents", "teams", "empty-role", "references")
		os.MkdirAll(roleDir, 0755)
		content := "name: empty-role\nskills: []\n"
		os.WriteFile(filepath.Join(roleDir, "role.yaml"), []byte(content), 0644)

		skills, err := ReadRoleSkills(dir, "empty-role")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(skills) != 0 {
			t.Errorf("expected 0 skills, got %d", len(skills))
		}
	})
}

func TestFindSkillPath(t *testing.T) {
	dir := t.TempDir()

	t.Run("not found returns empty", func(t *testing.T) {
		got := findSkillPath(dir, "nonexistent")
		if got != "" {
			t.Errorf("expected empty, got %q", got)
		}
	})

	t.Run("finds in .agents/teams/", func(t *testing.T) {
		teamDir := filepath.Join(dir, ".agents", "teams", "my-role")
		os.MkdirAll(teamDir, 0755)

		got := findSkillPath(dir, "my-role")
		if got != teamDir {
			t.Errorf("findSkillPath = %q, want %q", got, teamDir)
		}
	})

	t.Run("finds in skills/", func(t *testing.T) {
		skillDir := filepath.Join(dir, "skills", "my-skill")
		os.MkdirAll(skillDir, 0755)

		got := findSkillPath(dir, "my-skill")
		if got != skillDir {
			t.Errorf("findSkillPath = %q, want %q", got, skillDir)
		}
	})

	t.Run("finds in .claude/skills/", func(t *testing.T) {
		skillDir := filepath.Join(dir, ".claude", "skills", "local-cached-skill")
		os.MkdirAll(skillDir, 0755)

		got := findSkillPath(dir, "local-cached-skill")
		if got != skillDir {
			t.Errorf("findSkillPath(.claude/skills) = %q, want %q", got, skillDir)
		}
	})

	t.Run("scoped name resolves short name", func(t *testing.T) {
		skillDir := filepath.Join(dir, "skills", "vite")
		os.MkdirAll(skillDir, 0755)

		got := findSkillPath(dir, "antfu/skills@vite")
		if got != skillDir {
			t.Errorf("findSkillPath(antfu/skills@vite) = %q, want %q", got, skillDir)
		}
	})

	t.Run(".agents/teams takes priority over skills/", func(t *testing.T) {
		teamDir := filepath.Join(dir, ".agents", "teams", "dual")
		os.MkdirAll(teamDir, 0755)
		skillDir := filepath.Join(dir, "skills", "dual")
		os.MkdirAll(skillDir, 0755)

		got := findSkillPath(dir, "dual")
		if got != teamDir {
			t.Errorf("findSkillPath should prefer .agents/teams/, got %q", got)
		}
	})
}

func TestFindSkillPathPluginRoot(t *testing.T) {
	dir := t.TempDir()
	pluginRoot := t.TempDir()

	// Create skill in Plugin root
	pluginSkillDir := filepath.Join(pluginRoot, "skills", "plugin-builtin")
	os.MkdirAll(pluginSkillDir, 0755)
	os.WriteFile(filepath.Join(pluginSkillDir, "SKILL.md"), []byte("# plugin-builtin\n"), 0644)

	// Also create same skill in project to verify Plugin takes priority
	projSkillDir := filepath.Join(dir, "skills", "plugin-builtin")
	os.MkdirAll(projSkillDir, 0755)

	t.Setenv("CLAUDE_PLUGIN_ROOT", pluginRoot)

	got := findSkillPath(dir, "plugin-builtin")
	if got != pluginSkillDir {
		t.Errorf("Plugin root skill should take priority, got %q, want %q", got, pluginSkillDir)
	}
}

func TestCopySkillsToWorktree(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, "worktree")
	os.MkdirAll(wtPath, 0755)

	// Create role skill
	roleDir := filepath.Join(dir, ".agents", "teams", "dev")
	refDir := filepath.Join(roleDir, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# dev\n"), 0644)
	os.WriteFile(filepath.Join(roleDir, "system.md"), []byte("# System\n"), 0644)

	// Create dependency skill
	depDir := filepath.Join(dir, "skills", "vitest")
	os.MkdirAll(depDir, 0755)
	os.WriteFile(filepath.Join(depDir, "SKILL.md"), []byte("# vitest\n"), 0644)

	// Write role.yaml with skills
	roleYAML := "name: dev\nskills:\n  - vitest\n"
	os.WriteFile(filepath.Join(refDir, "role.yaml"), []byte(roleYAML), 0644)

	err := CopySkillsToWorktree(wtPath, dir, "dev")
	if err != nil {
		t.Fatalf("CopySkillsToWorktree: %v", err)
	}

	// Check role skill copied to .claude/skills/
	claudeSkill := filepath.Join(wtPath, ".claude", "skills", "dev", "SKILL.md")
	if _, err := os.Stat(claudeSkill); os.IsNotExist(err) {
		t.Error("role skill not copied to .claude/skills/")
	}

	// Check dependency skill copied
	claudeDep := filepath.Join(wtPath, ".claude", "skills", "vitest", "SKILL.md")
	if _, err := os.Stat(claudeDep); os.IsNotExist(err) {
		t.Error("dependency skill not copied to .claude/skills/")
	}

	// Check .codex/skills/ mirror
	codexSkill := filepath.Join(wtPath, ".codex", "skills", "dev", "SKILL.md")
	if _, err := os.Stat(codexSkill); os.IsNotExist(err) {
		t.Error("role skill not mirrored to .codex/skills/")
	}

	codexDep := filepath.Join(wtPath, ".codex", "skills", "vitest", "SKILL.md")
	if _, err := os.Stat(codexDep); os.IsNotExist(err) {
		t.Error("dependency skill not mirrored to .codex/skills/")
	}

	// Check .opencode/skills/ mirror
	opencodeSkill := filepath.Join(wtPath, ".opencode", "skills", "dev", "SKILL.md")
	if _, err := os.Stat(opencodeSkill); os.IsNotExist(err) {
		t.Error("role skill not mirrored to .opencode/skills/")
	}

	opencodeDep := filepath.Join(wtPath, ".opencode", "skills", "vitest", "SKILL.md")
	if _, err := os.Stat(opencodeDep); os.IsNotExist(err) {
		t.Error("dependency skill not mirrored to .opencode/skills/")
	}

	// Check .gemini/skills/ mirror
	geminiSkill := filepath.Join(wtPath, ".gemini", "skills", "dev", "SKILL.md")
	if _, err := os.Stat(geminiSkill); os.IsNotExist(err) {
		t.Error("role skill not mirrored to .gemini/skills/")
	}

	geminiDep := filepath.Join(wtPath, ".gemini", "skills", "vitest", "SKILL.md")
	if _, err := os.Stat(geminiDep); os.IsNotExist(err) {
		t.Error("dependency skill not mirrored to .gemini/skills/")
	}
}

func TestCopySkillsToWorktreeScopedName(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, "worktree")
	os.MkdirAll(wtPath, 0755)

	// Create role with scoped skill dependency
	roleDir := filepath.Join(dir, ".agents", "teams", "arch")
	refDir := filepath.Join(roleDir, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# arch\n"), 0644)

	// Create the skill that the scoped name resolves to
	viteDir := filepath.Join(dir, "skills", "vite")
	os.MkdirAll(viteDir, 0755)
	os.WriteFile(filepath.Join(viteDir, "SKILL.md"), []byte("# vite\n"), 0644)

	roleYAML := "name: arch\nskills:\n  - \"antfu/skills@vite\"\n"
	os.WriteFile(filepath.Join(refDir, "role.yaml"), []byte(roleYAML), 0644)

	err := CopySkillsToWorktree(wtPath, dir, "arch")
	if err != nil {
		t.Fatalf("CopySkillsToWorktree: %v", err)
	}

	// Should be copied as "vite", not "antfu/skills@vite"
	copiedSkill := filepath.Join(wtPath, ".claude", "skills", "vite", "SKILL.md")
	if _, err := os.Stat(copiedSkill); os.IsNotExist(err) {
		t.Error("scoped skill should be copied using short name 'vite'")
	}

	// Should NOT exist under the full scoped name
	badPath := filepath.Join(wtPath, ".claude", "skills", "antfu", "skills@vite")
	if _, err := os.Stat(badPath); err == nil {
		t.Error("scoped skill should NOT be copied using full scoped path")
	}
}

func TestProviderToAgent(t *testing.T) {
	tests := []struct {
		provider, want string
	}{
		{"claude", "claude-code"},
		{"codex", "codex"},
		{"opencode", "opencode"},
		{"gemini", "gemini"},
	}
	for _, tt := range tests {
		got, ok := providerToAgent[tt.provider]
		if !ok {
			t.Errorf("providerToAgent[%q] not found", tt.provider)
			continue
		}
		if got != tt.want {
			t.Errorf("providerToAgent[%q] = %q, want %q", tt.provider, got, tt.want)
		}
	}
}

func TestSkillTargetDir(t *testing.T) {
	tests := []struct {
		provider, wantSuffix string
	}{
		{"claude", filepath.Join(".claude", "skills")},
		{"codex", filepath.Join(".codex", "skills")},
		{"opencode", filepath.Join(".opencode", "skills")},
		{"gemini", filepath.Join(".gemini", "skills")},
		{"unknown", filepath.Join(".claude", "skills")},
	}
	for _, tt := range tests {
		got := skillTargetDir("/wt", tt.provider)
		want := filepath.Join("/wt", tt.wantSuffix)
		if got != want {
			t.Errorf("skillTargetDir(%q) = %q, want %q", tt.provider, got, want)
		}
	}
}

func TestIsScopedSkill(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"vite", false},
		{"antfu/skills@vite", true},
		{"jsonlee12138/prompts@eslint", true},
		{"plain-name", false},
	}
	for _, tt := range tests {
		got := isScopedSkill(tt.input)
		if got != tt.want {
			t.Errorf("isScopedSkill(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestInstallSkillsForWorkerLocalOnly(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, "worktree")
	os.MkdirAll(wtPath, 0755)

	// Create role skill
	roleDir := filepath.Join(dir, ".agents", "teams", "dev")
	refDir := filepath.Join(roleDir, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# dev\n"), 0644)

	// Create dependency skill in skills/
	depDir := filepath.Join(dir, "skills", "vitest")
	os.MkdirAll(depDir, 0755)
	os.WriteFile(filepath.Join(depDir, "SKILL.md"), []byte("# vitest\n"), 0644)

	// Write role.yaml with plain skills only (no scoped, so no npx calls)
	roleYAML := "name: dev\nskills:\n  - vitest\n"
	os.WriteFile(filepath.Join(refDir, "role.yaml"), []byte(roleYAML), 0644)

	err := InstallSkillsForWorker(wtPath, dir, "dev", "claude")
	if err != nil {
		t.Fatalf("InstallSkillsForWorker: %v", err)
	}

	// Check role skill copied to .claude/skills/
	claudeRole := filepath.Join(wtPath, ".claude", "skills", "dev", "SKILL.md")
	if _, err := os.Stat(claudeRole); os.IsNotExist(err) {
		t.Error("role skill not installed to .claude/skills/")
	}

	// Check dependency skill copied
	claudeDep := filepath.Join(wtPath, ".claude", "skills", "vitest", "SKILL.md")
	if _, err := os.Stat(claudeDep); os.IsNotExist(err) {
		t.Error("dependency skill not installed to .claude/skills/")
	}
}

func TestInstallSkillsForWorkerCodexProvider(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, "worktree")
	os.MkdirAll(wtPath, 0755)

	// Create role skill
	roleDir := filepath.Join(dir, ".agents", "teams", "dev")
	refDir := filepath.Join(roleDir, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# dev\n"), 0644)

	roleYAML := "name: dev\nskills: []\n"
	os.WriteFile(filepath.Join(refDir, "role.yaml"), []byte(roleYAML), 0644)

	err := InstallSkillsForWorker(wtPath, dir, "dev", "codex")
	if err != nil {
		t.Fatalf("InstallSkillsForWorker: %v", err)
	}

	// Check role skill copied to .codex/skills/
	codexRole := filepath.Join(wtPath, ".codex", "skills", "dev", "SKILL.md")
	if _, err := os.Stat(codexRole); os.IsNotExist(err) {
		t.Error("role skill not installed to .codex/skills/ for codex provider")
	}
}

func TestBridgeSkillsForProvider(t *testing.T) {
	dir := t.TempDir()

	// Create two roles with SKILL.md
	role1 := filepath.Join(dir, ".agents", "teams", "frontend-dev")
	os.MkdirAll(role1, 0755)
	os.WriteFile(filepath.Join(role1, "SKILL.md"), []byte("# frontend-dev\n"), 0644)

	role2 := filepath.Join(dir, ".agents", "teams", "backend-dev")
	os.MkdirAll(role2, 0755)
	os.WriteFile(filepath.Join(role2, "SKILL.md"), []byte("# backend-dev\n"), 0644)

	// Create a directory without SKILL.md (should not be bridged)
	noSkill := filepath.Join(dir, ".agents", "teams", "no-skill-dir")
	os.MkdirAll(noSkill, 0755)
	os.WriteFile(filepath.Join(noSkill, "README.md"), []byte("not a skill\n"), 0644)

	err := BridgeSkillsForProvider(dir, "claude")
	if err != nil {
		t.Fatalf("BridgeSkillsForProvider: %v", err)
	}

	// Check symlinks created
	link1 := filepath.Join(dir, ".claude", "skills", "frontend-dev")
	info1, err := os.Lstat(link1)
	if err != nil {
		t.Fatalf("symlink for frontend-dev not created: %v", err)
	}
	if info1.Mode()&os.ModeSymlink == 0 {
		t.Error("frontend-dev should be a symlink")
	}

	link2 := filepath.Join(dir, ".claude", "skills", "backend-dev")
	if _, err := os.Lstat(link2); err != nil {
		t.Fatalf("symlink for backend-dev not created: %v", err)
	}

	// Verify the symlink target is correct
	target, err := os.Readlink(link1)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if target != role1 {
		t.Errorf("symlink target = %q, want %q", target, role1)
	}

	// Check that no-skill-dir was NOT bridged
	noSkillLink := filepath.Join(dir, ".claude", "skills", "no-skill-dir")
	if _, err := os.Lstat(noSkillLink); !os.IsNotExist(err) {
		t.Error("directory without SKILL.md should not be bridged")
	}

	// Check SKILL.md is accessible through symlink
	data, err := os.ReadFile(filepath.Join(link1, "SKILL.md"))
	if err != nil {
		t.Fatalf("read through symlink: %v", err)
	}
	if string(data) != "# frontend-dev\n" {
		t.Errorf("SKILL.md content = %q", string(data))
	}
}

func TestBridgeSkillsForProviderSkipsExisting(t *testing.T) {
	dir := t.TempDir()

	// Create a role
	role := filepath.Join(dir, ".agents", "teams", "my-role")
	os.MkdirAll(role, 0755)
	os.WriteFile(filepath.Join(role, "SKILL.md"), []byte("# original\n"), 0644)

	// Pre-create a manual skill at the target (should NOT be overwritten)
	manualDir := filepath.Join(dir, ".claude", "skills", "my-role")
	os.MkdirAll(manualDir, 0755)
	os.WriteFile(filepath.Join(manualDir, "SKILL.md"), []byte("# manual install\n"), 0644)

	err := BridgeSkillsForProvider(dir, "claude")
	if err != nil {
		t.Fatalf("BridgeSkillsForProvider: %v", err)
	}

	// Should still be the manual version, not a symlink
	data, err := os.ReadFile(filepath.Join(manualDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "# manual install\n" {
		t.Errorf("existing skill was overwritten: %q", string(data))
	}
}

func TestBridgeSkillsForProviderOpenCode(t *testing.T) {
	dir := t.TempDir()

	role := filepath.Join(dir, ".agents", "teams", "dev")
	os.MkdirAll(role, 0755)
	os.WriteFile(filepath.Join(role, "SKILL.md"), []byte("# dev\n"), 0644)

	err := BridgeSkillsForProvider(dir, "opencode")
	if err != nil {
		t.Fatalf("BridgeSkillsForProvider: %v", err)
	}

	link := filepath.Join(dir, ".opencode", "skills", "dev")
	if _, err := os.Lstat(link); err != nil {
		t.Fatalf("symlink for opencode not created: %v", err)
	}
}
