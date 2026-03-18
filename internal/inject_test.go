// internal/inject_test.go
package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- 3B-4: Worker injection verification ---

// TestInjectRolePromptWithPath_SlimMode verifies that when .agents/rules/ exists,
// the injected CLAUDE.md uses slim mode (minimal content, rules reference).
func TestInjectRolePromptWithPath_SlimMode(t *testing.T) {
	root := t.TempDir()

	// Create .agents/rules/ with index.md
	rulesDir := filepath.Join(root, ".agents", "rules")
	os.MkdirAll(rulesDir, 0755)
	os.WriteFile(filepath.Join(rulesDir, "index.md"),
		[]byte("# Rules Index\n\n- `debugging.md`: bugs\n- `build-verification.md`: builds\n"), 0644)

	// Create role with system.md
	rolePath := filepath.Join(root, ".agents", "teams", "test-dev")
	refDir := filepath.Join(rolePath, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(rolePath, "system.md"),
		[]byte("# System Prompt: test-dev\n\nYou are the test-dev role.\n"), 0644)
	os.WriteFile(filepath.Join(rolePath, "SKILL.md"),
		[]byte("---\nname: test-dev\ndescription: Test development role\n---\n# test-dev\n"), 0644)
	os.WriteFile(filepath.Join(refDir, "role.yaml"),
		[]byte("name: test-dev\ndescription: \"Test development role\"\nskills:\n  - \"systematic-debugging\"\n"), 0644)

	// Create worktree dir
	wtPath := filepath.Join(root, ".worktrees", "test-dev-001")
	os.MkdirAll(wtPath, 0755)

	// Inject
	err := InjectRolePromptWithPath(wtPath, "test-dev-001", "test-dev", rolePath, root)
	if err != nil {
		t.Fatalf("InjectRolePromptWithPath: %v", err)
	}

	// Read injected CLAUDE.md
	data, err := os.ReadFile(filepath.Join(wtPath, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(data)

	// Must contain slim mode elements
	mustContain := []string{
		"AGENT_TEAM:START",
		"AGENT_TEAM:END",
		"test-dev",
		"Skill-First Workflow",
		"team/test-dev-001",
		"Rules Reference",
		".agents/rules/index.md",
	}
	for _, s := range mustContain {
		if !strings.Contains(content, s) {
			t.Errorf("CLAUDE.md missing expected content: %q", s)
		}
	}

	// Slim mode should NOT contain legacy inline Git Rules or Task Completion Protocol
	mustNotContain := []string{
		"git checkout",
		"git switch",
		"Task Completion Protocol",
		"git add -A",
		"agent-team task archive",
	}
	for _, s := range mustNotContain {
		if strings.Contains(content, s) {
			t.Errorf("CLAUDE.md in slim mode should NOT contain: %q", s)
		}
	}

	// Verify AGENTS.md and GEMINI.md also created
	for _, name := range []string{"AGENTS.md", "GEMINI.md"} {
		if _, err := os.Stat(filepath.Join(wtPath, name)); os.IsNotExist(err) {
			t.Errorf("%s should be created", name)
		}
	}
}

// TestInjectRolePromptWithPath_LegacyMode verifies that when .agents/rules/ does NOT exist,
// the full legacy template with inline Git Rules is used.
func TestInjectRolePromptWithPath_LegacyMode(t *testing.T) {
	root := t.TempDir()

	// No .agents/rules/ directory — legacy mode

	// Create role with system.md
	rolePath := filepath.Join(root, ".agents", "teams", "legacy-dev")
	refDir := filepath.Join(rolePath, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(rolePath, "system.md"),
		[]byte("# System Prompt: legacy-dev\n\nYou are the legacy-dev role.\n"), 0644)
	os.WriteFile(filepath.Join(refDir, "role.yaml"),
		[]byte("name: legacy-dev\ndescription: \"Legacy dev role\"\n"), 0644)

	// Create worktree dir
	wtPath := filepath.Join(root, ".worktrees", "legacy-dev-001")
	os.MkdirAll(wtPath, 0755)

	err := InjectRolePromptWithPath(wtPath, "legacy-dev-001", "legacy-dev", rolePath, root)
	if err != nil {
		t.Fatalf("InjectRolePromptWithPath: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(wtPath, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(data)

	// Legacy mode should contain inline Git Rules and Task Completion Protocol
	legacyStrings := []string{
		"Git Rules",
		"Task Completion Protocol",
		"git checkout",
		"agent-team task archive",
		"agent-team reply-main",
		"Skill-First Workflow",
	}
	for _, s := range legacyStrings {
		if !strings.Contains(content, s) {
			t.Errorf("CLAUDE.md in legacy mode missing: %q", s)
		}
	}

	// Legacy mode should NOT have Rules Reference section (no rules/ dir)
	if strings.Contains(content, "Rules Reference") {
		t.Error("CLAUDE.md in legacy mode should NOT contain Rules Reference")
	}
}

// TestInjectRolePromptWithPath_NoSystemMd verifies that if system.md is missing,
// nothing is injected.
func TestInjectRolePromptWithPath_NoSystemMd(t *testing.T) {
	root := t.TempDir()

	rolePath := filepath.Join(root, ".agents", "teams", "empty-role")
	os.MkdirAll(rolePath, 0755)
	// No system.md

	wtPath := filepath.Join(root, ".worktrees", "empty-role-001")
	os.MkdirAll(wtPath, 0755)

	err := InjectRolePromptWithPath(wtPath, "empty-role-001", "empty-role", rolePath, root)
	if err != nil {
		t.Fatalf("InjectRolePromptWithPath: %v", err)
	}

	// CLAUDE.md should NOT be created (nothing to inject)
	if _, err := os.Stat(filepath.Join(wtPath, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Error("CLAUDE.md should not be created when system.md is missing")
	}
}

// TestInjectRolePromptWithPath_RulesIndexInContent verifies that rules/index.md
// content appears in the injection output.
func TestInjectRolePromptWithPath_RulesIndexInContent(t *testing.T) {
	root := t.TempDir()

	rulesDir := filepath.Join(root, ".agents", "rules")
	os.MkdirAll(rulesDir, 0755)
	indexContent := "# Custom Rules\n\n- `my-rule.md`: custom stuff\n- `other-rule.md`: other things\n"
	os.WriteFile(filepath.Join(rulesDir, "index.md"), []byte(indexContent), 0644)

	rolePath := filepath.Join(root, ".agents", "teams", "my-role")
	refDir := filepath.Join(rolePath, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(rolePath, "system.md"),
		[]byte("# System Prompt: my-role\n"), 0644)
	os.WriteFile(filepath.Join(refDir, "role.yaml"),
		[]byte("name: my-role\n"), 0644)

	wtPath := filepath.Join(root, ".worktrees", "my-role-001")
	os.MkdirAll(wtPath, 0755)

	InjectRolePromptWithPath(wtPath, "my-role-001", "my-role", rolePath, root)

	data, _ := os.ReadFile(filepath.Join(wtPath, "CLAUDE.md"))
	content := string(data)

	if !strings.Contains(content, "my-rule.md") {
		t.Error("injected content should contain rules from index.md")
	}
	if !strings.Contains(content, "other-rule.md") {
		t.Error("injected content should contain all rules from index.md")
	}
}

// TestInjectRolePromptWithPath_DependencySkillsListed verifies that dependency skills
// from role.yaml appear in the injected content.
func TestInjectRolePromptWithPath_DependencySkillsListed(t *testing.T) {
	root := t.TempDir()

	rulesDir := filepath.Join(root, ".agents", "rules")
	os.MkdirAll(rulesDir, 0755)
	os.WriteFile(filepath.Join(rulesDir, "index.md"), []byte("# Rules\n"), 0644)

	rolePath := filepath.Join(root, ".agents", "teams", "full-role")
	refDir := filepath.Join(rolePath, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(rolePath, "system.md"),
		[]byte("# System Prompt: full-role\n"), 0644)
	os.WriteFile(filepath.Join(refDir, "role.yaml"),
		[]byte("name: full-role\nskills:\n  - \"git-commit\"\n  - \"systematic-debugging\"\n"), 0644)

	wtPath := filepath.Join(root, ".worktrees", "full-role-001")
	os.MkdirAll(wtPath, 0755)

	InjectRolePromptWithPath(wtPath, "full-role-001", "full-role", rolePath, root)

	data, _ := os.ReadFile(filepath.Join(wtPath, "CLAUDE.md"))
	content := string(data)

	if !strings.Contains(content, "git-commit") {
		t.Error("dependency skill 'git-commit' should appear in injected content")
	}
	if !strings.Contains(content, "systematic-debugging") {
		t.Error("dependency skill 'systematic-debugging' should appear in injected content")
	}
}

// TestHasRulesDir verifies the HasRulesDir helper.
func TestHasRulesDir(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agents", "rules"), 0755)
		if !HasRulesDir(dir) {
			t.Error("HasRulesDir should return true when .agents/rules/ exists")
		}
	})

	t.Run("not exists", func(t *testing.T) {
		dir := t.TempDir()
		if HasRulesDir(dir) {
			t.Error("HasRulesDir should return false when .agents/rules/ does not exist")
		}
	})

	t.Run("file not dir", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agents"), 0755)
		os.WriteFile(filepath.Join(dir, ".agents", "rules"), []byte("not a dir"), 0644)
		if HasRulesDir(dir) {
			t.Error("HasRulesDir should return false when rules is a file not a dir")
		}
	})
}

// TestInjectSection_BasicOperations verifies the generic InjectSection function.
func TestInjectSection_BasicOperations(t *testing.T) {
	t.Run("creates new file", func(t *testing.T) {
		fp := filepath.Join(t.TempDir(), "NEW.md")
		err := InjectSection(fp, "TEST", "hello world")
		if err != nil {
			t.Fatalf("InjectSection: %v", err)
		}
		data, _ := os.ReadFile(fp)
		content := string(data)
		if !strings.Contains(content, "<!-- TEST:START -->") {
			t.Error("missing start marker")
		}
		if !strings.Contains(content, "hello world") {
			t.Error("missing content")
		}
		if !strings.Contains(content, "<!-- TEST:END -->") {
			t.Error("missing end marker")
		}
	})

	t.Run("replaces existing section", func(t *testing.T) {
		fp := filepath.Join(t.TempDir(), "EXIST.md")
		InjectSection(fp, "TEST", "v1 content")
		InjectSection(fp, "TEST", "v2 content")

		data, _ := os.ReadFile(fp)
		content := string(data)
		if strings.Contains(content, "v1 content") {
			t.Error("v1 content should be replaced")
		}
		if !strings.Contains(content, "v2 content") {
			t.Error("v2 content should be present")
		}
		if strings.Count(content, "<!-- TEST:START -->") != 1 {
			t.Error("should have exactly one start marker")
		}
	})

	t.Run("preserves existing file content", func(t *testing.T) {
		fp := filepath.Join(t.TempDir(), "USER.md")
		os.WriteFile(fp, []byte("# My Notes\n\nKeep this!\n"), 0644)
		InjectSection(fp, "TEST", "injected")

		data, _ := os.ReadFile(fp)
		content := string(data)
		if !strings.Contains(content, "My Notes") {
			t.Error("user content should be preserved")
		}
		if !strings.Contains(content, "Keep this!") {
			t.Error("user content detail should be preserved")
		}
		if !strings.Contains(content, "injected") {
			t.Error("injected content should be present")
		}
	})
}
