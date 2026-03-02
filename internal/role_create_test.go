// internal/role_create_test.go
package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// helpers

func newTestConfig(roleName string) RoleConfig {
	return RoleConfig{
		RoleName:    roleName,
		Description: "Frontend role for UI implementation",
		SystemGoal:  "Ship accessible and maintainable UI work",
		InScope:     []string{"Implement UI components", "Improve page accessibility"},
		OutOfScope:  []string{"Database migrations", "Backend API ownership"},
		Skills:      []string{"vitest", "ui-ux-pro-max"},
	}
}

func fixedNow() time.Time { return time.Date(2026, 2, 25, 12, 0, 0, 0, time.UTC) }

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", path)
	}
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected file to NOT exist: %s", path)
	}
}

func assertDirNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected directory to NOT exist: %s", path)
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(b) != want {
		t.Errorf("file %s content = %q, want %q", path, string(b), want)
	}
}

func assertFileContains(t *testing.T, path, substr string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(b), substr) {
		t.Errorf("file %s does not contain %q", path, substr)
	}
}

func assertStringSlice(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("slice len %d != %d; got=%v want=%v", len(got), len(want), got, want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("slice[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// --- Test #1: ValidateRoleName ---

func TestValidateRoleName_InvalidIncludesSuggestion(t *testing.T) {
	_, err := ValidateRoleName("FrontEnd Dev")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "frontend-dev") {
		t.Errorf("error should contain suggestion 'frontend-dev', got: %s", err.Error())
	}
}

// --- Test #2: RenderFiles deterministic ---

func TestRenderFiles_Deterministic(t *testing.T) {
	config := newTestConfig("frontend-dev")
	first, err := RenderFiles(config)
	if err != nil {
		t.Fatal(err)
	}
	second, err := RenderFiles(config)
	if err != nil {
		t.Fatal(err)
	}
	for path, content := range first {
		if second[path] != content {
			t.Errorf("non-deterministic output for %s", path)
		}
	}
}

// --- Test #3: RenderFiles system.md includes skill policy ---

func TestRenderFiles_SystemPromptIncludesSkillPolicy(t *testing.T) {
	config := RoleConfig{
		RoleName:    "frontend-dev",
		Description: "Frontend role for UI implementation",
		SystemGoal:  "Ship accessible and maintainable UI work",
		InScope:     []string{"Implement UI components"},
		OutOfScope:  []string{"Backend API ownership"},
		Skills:      []string{"vitest"},
	}
	rendered, err := RenderFiles(config)
	if err != nil {
		t.Fatal(err)
	}
	systemMD := rendered["system.md"]
	checks := []string{
		"If a required skill is missing at runtime, use `find-skills` to recommend installable skills for this role.",
		"Before any installation, ask the user whether to install globally or project-level.",
		"If the user does not specify, default to global installation.",
	}
	for _, want := range checks {
		if !strings.Contains(systemMD, want) {
			t.Errorf("system.md missing expected text: %q", want)
		}
	}
}

// --- Test #4: CreateOrUpdateRole overwrite creates backup ---

func TestCreateOrUpdateRole_OverwriteCreatesBackup(t *testing.T) {
	repoRoot := t.TempDir()
	target := filepath.Join(repoRoot, "skills", "frontend-dev")
	refs := filepath.Join(target, "references")
	mkdirAll(t, refs)
	writeFile(t, filepath.Join(target, "SKILL.md"), "old skill\n")
	writeFile(t, filepath.Join(target, "role.yaml"), "legacy root yaml\n")
	writeFile(t, filepath.Join(refs, "role.yaml"), "old yaml\n")
	writeFile(t, filepath.Join(target, "system.md"), "old system\n")
	writeFile(t, filepath.Join(target, "keep.txt"), "keep me\n")

	config := RoleConfig{
		RoleName:    "frontend-dev",
		Description: "Frontend role for UI implementation",
		SystemGoal:  "Ship accessible and maintainable UI work",
		InScope:     []string{"Implement UI components"},
		OutOfScope:  []string{"Own backend services"},
		Skills:      []string{"vitest"},
	}

	result, err := CreateOrUpdateRole(repoRoot, config, "yes", nil, fixedNow, "skills")
	if err != nil {
		t.Fatal(err)
	}

	expectedBackup := filepath.Join(repoRoot, "skills", ".backup", "frontend-dev-20260225-120000")
	if result.BackupPath != expectedBackup {
		t.Errorf("BackupPath = %q, want %q", result.BackupPath, expectedBackup)
	}
	assertFileExists(t, filepath.Join(expectedBackup, "SKILL.md"))
	assertFileContent(t, filepath.Join(expectedBackup, "references", "role.yaml"), "old yaml\n")
	assertFileContent(t, filepath.Join(expectedBackup, "role.yaml"), "legacy root yaml\n")
	assertFileContent(t, filepath.Join(expectedBackup, "SKILL.md"), "old skill\n")
	assertFileContent(t, filepath.Join(target, "keep.txt"), "keep me\n")
	assertFileContains(t, filepath.Join(target, "SKILL.md"), "Frontend role for UI implementation")
	assertFileContains(t, filepath.Join(target, "references", "role.yaml"), "Frontend role for UI implementation")
	assertFileNotExists(t, filepath.Join(target, "role.yaml"))
}

// --- Test #5: ResolveFinalSkills manual fallback ---

func TestResolveFinalSkills_ManualFallback(t *testing.T) {
	final := ResolveFinalSkills(nil, nil, nil, nil, []string{"custom-skill-a", "custom-skill-b"})
	assertStringSlice(t, final, []string{"custom-skill-a", "custom-skill-b"})
}

// --- Test #6: ResolveFinalSkills empty allowed ---

func TestResolveFinalSkills_EmptyAllowed(t *testing.T) {
	final := ResolveFinalSkills(nil, nil, nil, nil, nil)
	if len(final) != 0 {
		t.Errorf("expected empty slice, got %v", final)
	}
}

// --- Test #7: RenderFiles empty skills = "skills: []" ---

func TestRenderFiles_EmptySkillsYAMLArray(t *testing.T) {
	config := RoleConfig{
		RoleName:    "frontend-dev",
		Description: "Frontend role",
		SystemGoal:  "Ship UI",
		InScope:     []string{"Build components"},
		OutOfScope:  []string{"Own backend"},
		Skills:      []string{},
	}
	rendered, err := RenderFiles(config)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered["references/role.yaml"], "skills: []") {
		t.Errorf("expected 'skills: []' in role.yaml, got:\n%s", rendered["references/role.yaml"])
	}
}

// --- Test #8: RenderFiles non-empty skills = list shape ---

func TestRenderFiles_NonEmptySkillsYAMLList(t *testing.T) {
	config := RoleConfig{
		RoleName:    "frontend-dev",
		Description: "Frontend role",
		SystemGoal:  "Ship UI",
		InScope:     []string{"Build components"},
		OutOfScope:  []string{"Own backend"},
		Skills:      []string{"vitest", "ui-ux-pro-max"},
	}
	rendered, err := RenderFiles(config)
	if err != nil {
		t.Fatal(err)
	}
	want := "skills:\n  - \"vitest\"\n  - \"ui-ux-pro-max\""
	if !strings.Contains(rendered["references/role.yaml"], want) {
		t.Errorf("expected skills list in role.yaml:\n%s\ngot:\n%s", want, rendered["references/role.yaml"])
	}
}

// --- Test #9: ParseSelectionReply checkbox mode ---

func TestParseSelectionReply_CheckboxMode(t *testing.T) {
	recommended := []string{"ui-ux-pro-max", "vitest", "better-icons"}
	reply := "1. [x] ui-ux-pro-max\n2. [ ] vitest\n3. [x] better-icons"
	selected, mode, precedence := ParseSelectionReply(reply, recommended)
	assertStringSlice(t, selected, []string{"ui-ux-pro-max", "better-icons"})
	if mode != "checkbox" {
		t.Errorf("mode = %q, want checkbox", mode)
	}
	if precedence {
		t.Error("checkbox_precedence should be false")
	}
}

// --- Test #10: ParseSelectionReply numeric mode ---

func TestParseSelectionReply_NumericMode(t *testing.T) {
	recommended := []string{"ui-ux-pro-max", "vitest", "better-icons"}
	selected, mode, precedence := ParseSelectionReply("1,3", recommended)
	assertStringSlice(t, selected, []string{"ui-ux-pro-max", "better-icons"})
	if mode != "numeric" {
		t.Errorf("mode = %q, want numeric", mode)
	}
	if precedence {
		t.Error("checkbox_precedence should be false")
	}
}

// --- Test #11: ParseSelectionReply checkbox precedence ---

func TestParseSelectionReply_CheckboxPrecedence(t *testing.T) {
	recommended := []string{"ui-ux-pro-max", "vitest", "better-icons"}
	reply := "1. [ ] ui-ux-pro-max\n2. [x] vitest\n3. [ ] better-icons\n1,3"
	selected, mode, precedence := ParseSelectionReply(reply, recommended)
	assertStringSlice(t, selected, []string{"vitest"})
	if mode != "checkbox" {
		t.Errorf("mode = %q, want checkbox", mode)
	}
	if !precedence {
		t.Error("checkbox_precedence should be true")
	}
}

// --- Test #12: CreateOrUpdateRole happy path ---

func TestCreateOrUpdateRole_HappyPath(t *testing.T) {
	repoRoot := t.TempDir()
	config := RoleConfig{
		RoleName:    "data-engineer",
		Description: "Data pipeline and ETL role",
		SystemGoal:  "Build reliable data pipelines",
		InScope:     []string{"ETL jobs", "Data validation"},
		OutOfScope:  []string{"Frontend work"},
		Skills:      []string{"vitest"},
	}
	result, err := CreateOrUpdateRole(repoRoot, config, "ask", nil, fixedNow, "skills")
	if err != nil {
		t.Fatal(err)
	}
	if result.BackupPath != "" {
		t.Errorf("expected no backup, got %q", result.BackupPath)
	}
	target := filepath.Join(repoRoot, "skills", "data-engineer")
	assertFileExists(t, filepath.Join(target, "SKILL.md"))
	assertFileExists(t, filepath.Join(target, "references", "role.yaml"))
	assertFileExists(t, filepath.Join(target, "system.md"))
	assertFileContains(t, filepath.Join(target, "SKILL.md"), "data-engineer")
	assertFileContains(t, filepath.Join(target, "SKILL.md"), "Data pipeline and ETL role")
	assertFileContains(t, filepath.Join(target, "references", "role.yaml"), "ETL jobs")
	assertFileContains(t, filepath.Join(target, "references", "role.yaml"), "Frontend work")
}

// --- Test #13: CreateOrUpdateRole overwrite=no raises error ---

func TestCreateOrUpdateRole_OverwriteModeNoError(t *testing.T) {
	repoRoot := t.TempDir()
	target := filepath.Join(repoRoot, "skills", "existing-role")
	mkdirAll(t, target)
	writeFile(t, filepath.Join(target, "SKILL.md"), "old\n")

	config := RoleConfig{
		RoleName:    "existing-role",
		Description: "Some role",
		SystemGoal:  "Do stuff",
		InScope:     []string{"Task A"},
		OutOfScope:  []string{"Task B"},
		Skills:      []string{},
	}
	_, err := CreateOrUpdateRole(repoRoot, config, "no", nil, fixedNow, "skills")
	if err == nil {
		t.Fatal("expected error for overwrite=no")
	}
}

// --- Test #14: NormalizeRoleName edge cases ---

func TestNormalizeRoleName_EdgeCases(t *testing.T) {
	cases := []struct{ in, want string }{
		{"  FrontEnd Dev  ", "frontend-dev"},
		{"a---b", "a-b"},
		{"Hello World!@#", "hello-world"},
		{"   ", ""},
		{"UPPER_CASE_NAME", "upper-case-name"},
		{"already-kebab", "already-kebab"},
	}
	for _, c := range cases {
		got := NormalizeRoleName(c.in)
		if got != c.want {
			t.Errorf("NormalizeRoleName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// --- Test #15: CollectScope fallback ---

func TestCollectScope_Fallback(t *testing.T) {
	got := CollectScope(nil, "Default scope")
	assertStringSlice(t, got, []string{"Default scope"})

	got = CollectScope([]string{"", "  "}, "Default scope")
	assertStringSlice(t, got, []string{"Default scope"})

	got = CollectScope([]string{"a,b", "c"}, "Default")
	assertStringSlice(t, got, []string{"a", "b", "c"})
}

// --- Test #16: CreateOrUpdateRole target dir variants ---

func TestCreateOrUpdateRole_TargetDirVariants(t *testing.T) {
	t.Run("agents_teams", func(t *testing.T) {
		repoRoot := t.TempDir()
		config := RoleConfig{
			RoleName:    "backend-dev",
			Description: "Backend development role",
			SystemGoal:  "Build reliable backend services",
			InScope:     []string{"API design", "Database queries"},
			OutOfScope:  []string{"Frontend work"},
			Skills:      []string{"vitest"},
		}
		result, err := CreateOrUpdateRole(repoRoot, config, "ask", nil, fixedNow, ".agents/teams")
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(repoRoot, ".agents", "teams", "backend-dev")
		if result.TargetDir != want {
			t.Errorf("TargetDir = %q, want %q", result.TargetDir, want)
		}
		assertFileExists(t, filepath.Join(want, "SKILL.md"))
		assertFileExists(t, filepath.Join(want, "references", "role.yaml"))
		assertFileExists(t, filepath.Join(want, "system.md"))
		assertDirNotExists(t, filepath.Join(repoRoot, "skills", "backend-dev"))
	})

	t.Run("custom_path", func(t *testing.T) {
		repoRoot := t.TempDir()
		config := RoleConfig{
			RoleName:    "custom-role",
			Description: "Custom role",
			SystemGoal:  "Do custom work",
			InScope:     []string{"Custom tasks"},
			OutOfScope:  []string{"Other tasks"},
			Skills:      []string{},
		}
		result, err := CreateOrUpdateRole(repoRoot, config, "ask", nil, fixedNow, "my-custom-dir")
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(repoRoot, "my-custom-dir", "custom-role")
		if result.TargetDir != want {
			t.Errorf("TargetDir = %q, want %q", result.TargetDir, want)
		}
		assertFileExists(t, filepath.Join(want, "SKILL.md"))
	})
}
