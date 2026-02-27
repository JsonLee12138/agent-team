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
		roleDir := filepath.Join(dir, "agents", "teams", "dev", "references")
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
		roleDir := filepath.Join(dir, "agents", "teams", "empty-role", "references")
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

	t.Run("finds in agents/teams/", func(t *testing.T) {
		teamDir := filepath.Join(dir, "agents", "teams", "my-role")
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

	t.Run("scoped name resolves short name", func(t *testing.T) {
		skillDir := filepath.Join(dir, "skills", "vite")
		os.MkdirAll(skillDir, 0755)

		got := findSkillPath(dir, "antfu/skills@vite")
		if got != skillDir {
			t.Errorf("findSkillPath(antfu/skills@vite) = %q, want %q", got, skillDir)
		}
	})

	t.Run("agents/teams takes priority over skills/", func(t *testing.T) {
		teamDir := filepath.Join(dir, "agents", "teams", "dual")
		os.MkdirAll(teamDir, 0755)
		skillDir := filepath.Join(dir, "skills", "dual")
		os.MkdirAll(skillDir, 0755)

		got := findSkillPath(dir, "dual")
		if got != teamDir {
			t.Errorf("findSkillPath should prefer agents/teams/, got %q", got)
		}
	})
}

func TestCopySkillsToWorktree(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, "worktree")
	os.MkdirAll(wtPath, 0755)

	// Create role skill
	roleDir := filepath.Join(dir, "agents", "teams", "dev")
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
}

func TestCopySkillsToWorktreeScopedName(t *testing.T) {
	dir := t.TempDir()
	wtPath := filepath.Join(dir, "worktree")
	os.MkdirAll(wtPath, 0755)

	// Create role with scoped skill dependency
	roleDir := filepath.Join(dir, "agents", "teams", "arch")
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
