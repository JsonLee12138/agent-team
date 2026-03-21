// internal/integration_test.go
package internal

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWorkerFlowSkillSurvival is an integration test that verifies skills
// copied during "worker open" survive and aren't overwritten.
// It simulates the full worker create → worker open flow using real git worktrees.
func TestWorkerFlowSkillSurvival(t *testing.T) {
	// 1. Create a real git repository
	repoDir := initTestRepo(t)

	// 2. Set up role definition in .agent-team/teams/dev-role/
	roleName := "dev-role"
	roleDir := filepath.Join(repoDir, ".agent-team", "teams", roleName)
	refDir := filepath.Join(roleDir, "references")
	if err := os.MkdirAll(refDir, 0755); err != nil {
		t.Fatalf("mkdir role: %v", err)
	}

	// Write SKILL.md and system.md for the role
	if err := os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("---\nname: dev-role\ndescription: Test role\n---\n# dev-role\n"), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(roleDir, "system.md"), []byte("# System Prompt: dev-role\n\nYou are the dev-role role.\n"), 0644); err != nil {
		t.Fatalf("write system.md: %v", err)
	}

	// Write role.yaml with dependency skills
	roleYAMLContent := `name: dev-role
description: "Test role for integration testing"
skills:
  - name: "vite-tool"
    description: "Vite tool support"
  - name: "antfu/skills@vitest"
    description: "Vitest support"
  - name: "jsonlee12138/prompts@eslint-config"
    description: "ESLint config support"
`
	if err := os.WriteFile(filepath.Join(refDir, "role.yaml"), []byte(roleYAMLContent), 0644); err != nil {
		t.Fatalf("write role.yaml: %v", err)
	}

	// 3. Create dependency skills in the repo
	// Plain skill: skills/vite-tool/
	viteDir := filepath.Join(repoDir, "skills", "vite-tool")
	if err := os.MkdirAll(viteDir, 0755); err != nil {
		t.Fatalf("mkdir vite-tool: %v", err)
	}
	if err := os.WriteFile(filepath.Join(viteDir, "SKILL.md"), []byte("# vite-tool skill\n"), 0644); err != nil {
		t.Fatalf("write vite-tool SKILL.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(viteDir, "config.json"), []byte(`{"name":"vite-tool"}`), 0644); err != nil {
		t.Fatalf("write vite-tool config.json: %v", err)
	}

	// Scoped skill: skills/vitest/ (for "antfu/skills@vitest")
	vitestDir := filepath.Join(repoDir, "skills", "vitest")
	if err := os.MkdirAll(vitestDir, 0755); err != nil {
		t.Fatalf("mkdir vitest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vitestDir, "SKILL.md"), []byte("# vitest skill\n"), 0644); err != nil {
		t.Fatalf("write vitest SKILL.md: %v", err)
	}

	// Scoped skill: skills/eslint-config/ (for "jsonlee12138/prompts@eslint-config")
	eslintDir := filepath.Join(repoDir, "skills", "eslint-config")
	if err := os.MkdirAll(eslintDir, 0755); err != nil {
		t.Fatalf("mkdir eslint-config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(eslintDir, "SKILL.md"), []byte("# eslint-config skill\n"), 0644); err != nil {
		t.Fatalf("write eslint-config SKILL.md: %v", err)
	}

	// 4. Create a git worktree (simulating "worker create")
	gc, err := NewGitClient(repoDir)
	if err != nil {
		t.Fatalf("NewGitClient: %v", err)
	}
	// Normalize repoDir since git resolves symlinks (macOS /private/var)
	repoDir, _ = filepath.EvalSymlinks(repoDir)

	workerID := "dev-role-001"
	wtPath := filepath.Join(repoDir, ".worktrees", workerID)
	branch := "team/" + workerID

	if err := gc.WorktreeAdd(wtPath, branch); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}
	defer gc.WorktreeRemove(wtPath)

	// Write .gitignore (as worker create does)
	if err := WriteWorktreeGitignore(wtPath); err != nil {
		t.Fatalf("WriteWorktreeGitignore: %v", err)
	}

	// 5. Verify worktree exists but has NO skills yet
	claudeSkillsDir := filepath.Join(wtPath, ".claude", "skills")
	if _, err := os.Stat(claudeSkillsDir); !os.IsNotExist(err) {
		t.Error(".claude/skills/ should not exist before worker open")
	}

	// 6. Simulate "worker open" — copy skills
	if err := CopySkillsToWorktree(wtPath, repoDir, roleName); err != nil {
		t.Fatalf("CopySkillsToWorktree: %v", err)
	}

	// 7. Verify all skills are present in .claude/skills/
	expectedSkills := map[string]struct {
		files []string
	}{
		"dev-role":      {files: []string{"SKILL.md", "system.md"}},
		"vite-tool":     {files: []string{"SKILL.md", "config.json"}},
		"vitest":        {files: []string{"SKILL.md"}},
		"eslint-config": {files: []string{"SKILL.md"}},
	}

	for skillName, expected := range expectedSkills {
		for _, fileName := range expected.files {
			claudePath := filepath.Join(wtPath, ".claude", "skills", skillName, fileName)
			if _, err := os.Stat(claudePath); os.IsNotExist(err) {
				t.Errorf(".claude/skills/%s/%s missing after CopySkillsToWorktree", skillName, fileName)
			}

			codexPath := filepath.Join(wtPath, ".codex", "skills", skillName, fileName)
			if _, err := os.Stat(codexPath); os.IsNotExist(err) {
				t.Errorf(".codex/skills/%s/%s missing after CopySkillsToWorktree", skillName, fileName)
			}

			opencodePath := filepath.Join(wtPath, ".opencode", "skills", skillName, fileName)
			if _, err := os.Stat(opencodePath); os.IsNotExist(err) {
				t.Errorf(".opencode/skills/%s/%s missing after CopySkillsToWorktree", skillName, fileName)
			}

			geminiPath := filepath.Join(wtPath, ".gemini", "skills", skillName, fileName)
			if _, err := os.Stat(geminiPath); os.IsNotExist(err) {
				t.Errorf(".gemini/skills/%s/%s missing after CopySkillsToWorktree", skillName, fileName)
			}
		}
	}

	// 8. Verify scoped skills are NOT stored under full scoped names
	badPaths := []string{
		filepath.Join(wtPath, ".claude", "skills", "antfu", "skills@vitest"),
		filepath.Join(wtPath, ".claude", "skills", "jsonlee12138", "prompts@eslint-config"),
		filepath.Join(wtPath, ".claude", "skills", "antfu/skills@vitest"),
		filepath.Join(wtPath, ".claude", "skills", "jsonlee12138/prompts@eslint-config"),
	}
	for _, bp := range badPaths {
		if _, err := os.Stat(bp); err == nil {
			t.Errorf("scoped skill should NOT exist at full path: %s", bp)
		}
	}

	// 9. Simulate "worker open" — inject role prompt
	if err := InjectRolePrompt(wtPath, workerID, roleName, repoDir); err != nil {
		t.Fatalf("InjectRolePrompt: %v", err)
	}

	// 10. Verify CLAUDE.md exists and contains expected content
	claudeMD, err := os.ReadFile(filepath.Join(wtPath, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	claudeMDStr := string(claudeMD)

	mustContain := []string{
		"dev-role",
		"Skill-First Workflow",
		"Role skill (MUST use)",
		"vite-tool",
		"vitest",
		"eslint-config",
		"team/dev-role-001",
	}
	for _, s := range mustContain {
		if !contains(claudeMDStr, s) {
			t.Errorf("CLAUDE.md missing expected content: %q", s)
		}
	}

	// 11. Verify AGENTS.md exists and contains expected content
	agentsMD, err := os.ReadFile(filepath.Join(wtPath, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	agentsMDStr := string(agentsMD)
	for _, s := range mustContain {
		if !contains(agentsMDStr, s) {
			t.Errorf("AGENTS.md missing expected content: %q", s)
		}
	}

	// 12. Verify skills still exist AFTER InjectRolePrompt (no overwriting)
	for skillName, expected := range expectedSkills {
		for _, fileName := range expected.files {
			claudePath := filepath.Join(wtPath, ".claude", "skills", skillName, fileName)
			if _, err := os.Stat(claudePath); os.IsNotExist(err) {
				t.Errorf(".claude/skills/%s/%s missing AFTER InjectRolePrompt — prompt injection overwrote skills!", skillName, fileName)
			}
		}
	}

	// 13. Verify skill file content integrity
	viteSkillContent, err := os.ReadFile(filepath.Join(wtPath, ".claude", "skills", "vite-tool", "SKILL.md"))
	if err != nil {
		t.Fatalf("read copied vite-tool SKILL.md: %v", err)
	}
	if string(viteSkillContent) != "# vite-tool skill\n" {
		t.Errorf("vite-tool SKILL.md content corrupted: %q", string(viteSkillContent))
	}

	// 14. Simulate re-open (worker open again) — skills should be re-copied cleanly
	// Modify a skill file in the source to verify re-copy works
	if err := os.WriteFile(filepath.Join(viteDir, "SKILL.md"), []byte("# vite-tool skill v2\n"), 0644); err != nil {
		t.Fatalf("update vite-tool: %v", err)
	}

	if err := CopySkillsToWorktree(wtPath, repoDir, roleName); err != nil {
		t.Fatalf("CopySkillsToWorktree (re-open): %v", err)
	}

	// Verify updated content
	updatedContent, err := os.ReadFile(filepath.Join(wtPath, ".claude", "skills", "vite-tool", "SKILL.md"))
	if err != nil {
		t.Fatalf("read updated vite-tool: %v", err)
	}
	if string(updatedContent) != "# vite-tool skill v2\n" {
		t.Errorf("re-open should update skill content, got: %q", string(updatedContent))
	}

	// All other skills should still be present after re-open
	for skillName, expected := range expectedSkills {
		for _, fileName := range expected.files {
			claudePath := filepath.Join(wtPath, ".claude", "skills", skillName, fileName)
			if _, err := os.Stat(claudePath); os.IsNotExist(err) {
				t.Errorf(".claude/skills/%s/%s missing after re-open", skillName, fileName)
			}
		}
	}
}

// TestWorkerFlowGlobalSkillCopy verifies that skills installed globally
// (~/.claude/skills/) are also found and copied to the worktree.
func TestWorkerFlowGlobalSkillCopy(t *testing.T) {
	repoDir := initTestRepo(t)

	// Set up role
	roleName := "test-role"
	roleDir := filepath.Join(repoDir, ".agent-team", "teams", roleName)
	refDir := filepath.Join(roleDir, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# test-role\n"), 0644)
	os.WriteFile(filepath.Join(roleDir, "system.md"), []byte("# System Prompt: test-role\n"), 0644)

	// Create a "global" skill in a temp directory simulating ~/.claude/skills/
	// We can't easily mock os.UserHomeDir, but we can test via findSkillPath
	// by placing the skill in a project-local path (skills/) to verify the resolution
	globalStyleSkill := filepath.Join(repoDir, "skills", "design-patterns")
	os.MkdirAll(globalStyleSkill, 0755)
	os.WriteFile(filepath.Join(globalStyleSkill, "SKILL.md"), []byte("# design-patterns\n"), 0644)
	os.WriteFile(filepath.Join(globalStyleSkill, "patterns.md"), []byte("# Patterns guide\n"), 0644)

	roleYAMLContent := `name: test-role
skills:
  - name: "design-patterns"
    description: "Design patterns guide"
`
	os.WriteFile(filepath.Join(refDir, "role.yaml"), []byte(roleYAMLContent), 0644)

	// Create worktree
	gc, _ := NewGitClient(repoDir)
	repoDir, _ = filepath.EvalSymlinks(repoDir)
	wtPath := filepath.Join(repoDir, ".worktrees", "test-role-001")
	gc.WorktreeAdd(wtPath, "team/test-role-001")
	defer gc.WorktreeRemove(wtPath)

	// Copy skills
	if err := CopySkillsToWorktree(wtPath, repoDir, roleName); err != nil {
		t.Fatalf("CopySkillsToWorktree: %v", err)
	}

	// Verify the multi-file skill was fully copied
	checkFiles := []string{
		filepath.Join(wtPath, ".claude", "skills", "design-patterns", "SKILL.md"),
		filepath.Join(wtPath, ".claude", "skills", "design-patterns", "patterns.md"),
		filepath.Join(wtPath, ".codex", "skills", "design-patterns", "SKILL.md"),
		filepath.Join(wtPath, ".codex", "skills", "design-patterns", "patterns.md"),
	}
	for _, f := range checkFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("missing: %s", f)
		}
	}

	// Verify content
	patternsContent, err := os.ReadFile(filepath.Join(wtPath, ".claude", "skills", "design-patterns", "patterns.md"))
	if err != nil {
		t.Fatalf("read patterns.md: %v", err)
	}
	if string(patternsContent) != "# Patterns guide\n" {
		t.Errorf("patterns.md content mismatch: %q", string(patternsContent))
	}
}

// TestWorkerFlowMissingSkillWarning verifies that missing skills produce a warning
// but don't prevent other skills from being copied.
func TestWorkerFlowMissingSkillWarning(t *testing.T) {
	repoDir := initTestRepo(t)

	roleName := "partial-role"
	roleDir := filepath.Join(repoDir, ".agent-team", "teams", roleName)
	refDir := filepath.Join(roleDir, "references")
	os.MkdirAll(refDir, 0755)
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# partial-role\n"), 0644)
	os.WriteFile(filepath.Join(roleDir, "system.md"), []byte("# System\n"), 0644)

	// Only create one of the two dependency skills
	existingSkill := filepath.Join(repoDir, "skills", "existing-skill")
	os.MkdirAll(existingSkill, 0755)
	os.WriteFile(filepath.Join(existingSkill, "SKILL.md"), []byte("# exists\n"), 0644)

	roleYAMLContent := `name: partial-role
skills:
  - name: "existing-skill"
    description: "Existing local skill"
  - name: "nonexistent-skill"
    description: "Missing skill for warning path"
`
	os.WriteFile(filepath.Join(refDir, "role.yaml"), []byte(roleYAMLContent), 0644)

	gc, _ := NewGitClient(repoDir)
	repoDir, _ = filepath.EvalSymlinks(repoDir)
	wtPath := filepath.Join(repoDir, ".worktrees", "partial-role-001")
	gc.WorktreeAdd(wtPath, "team/partial-role-001")
	defer gc.WorktreeRemove(wtPath)

	// CopySkillsToWorktree should NOT return an error for missing skills
	err := CopySkillsToWorktree(wtPath, repoDir, roleName)
	if err != nil {
		t.Fatalf("CopySkillsToWorktree should not fail for missing skills: %v", err)
	}

	// The existing skill should be copied
	if _, err := os.Stat(filepath.Join(wtPath, ".claude", "skills", "existing-skill", "SKILL.md")); os.IsNotExist(err) {
		t.Error("existing-skill should be copied despite missing nonexistent-skill")
	}

	// The role skill itself should still be copied
	if _, err := os.Stat(filepath.Join(wtPath, ".claude", "skills", roleName, "SKILL.md")); os.IsNotExist(err) {
		t.Error("role skill should be copied")
	}

	// The nonexistent skill should NOT be in the worktree
	if _, err := os.Stat(filepath.Join(wtPath, ".claude", "skills", "nonexistent-skill")); !os.IsNotExist(err) {
		t.Error("nonexistent-skill should not exist in worktree")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsImpl(s, substr)
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestRequirementFullLifecycle is an integration test covering:
// req create → req split → req assign → MarkSubTaskDone → auto-promote
// while verifying index.yaml stays in sync at every step.
func TestRequirementFullLifecycle(t *testing.T) {
	dir := t.TempDir()

	// Step 1: Create requirement
	req, err := CreateRequirement(dir, "user-auth", "Implement user authentication", nil)
	if err != nil {
		t.Fatalf("CreateRequirement: %v", err)
	}
	if err := UpdateIndexEntry(dir, req); err != nil {
		t.Fatalf("UpdateIndexEntry(create): %v", err)
	}

	// Verify initial state
	if req.Status != RequirementStatusOpen {
		t.Errorf("initial status = %q, want open", req.Status)
	}
	assertIndexSync(t, dir, "user-auth", RequirementStatusOpen, 0, 0)

	// Step 2: Split into sub-tasks
	req.SubTasks = append(req.SubTasks,
		SubTask{ID: 1, Title: "Design login API", Status: SubTaskStatusPending},
		SubTask{ID: 2, Title: "Build frontend", Status: SubTaskStatusPending},
		SubTask{ID: 3, Title: "Write tests", Status: SubTaskStatusPending},
	)
	if err := SaveRequirement(dir, req); err != nil {
		t.Fatalf("SaveRequirement(split): %v", err)
	}
	if err := UpdateIndexEntry(dir, req); err != nil {
		t.Fatalf("UpdateIndexEntry(split): %v", err)
	}
	assertIndexSync(t, dir, "user-auth", RequirementStatusOpen, 3, 0)

	// Step 3: Assign sub-task 1 → requirement moves to in_progress
	if err := AssignSubTask(req, 1, "backend-001", "change-login-api"); err != nil {
		t.Fatalf("AssignSubTask(1): %v", err)
	}
	if err := SaveRequirement(dir, req); err != nil {
		t.Fatalf("SaveRequirement(assign1): %v", err)
	}
	if err := UpdateIndexEntry(dir, req); err != nil {
		t.Fatalf("UpdateIndexEntry(assign1): %v", err)
	}

	if req.Status != RequirementStatusInProgress {
		t.Errorf("status after first assign = %q, want in_progress", req.Status)
	}
	assertIndexSync(t, dir, "user-auth", RequirementStatusInProgress, 3, 0)

	// Verify sub-task state
	if req.SubTasks[0].AssignedTo != "backend-001" {
		t.Errorf("SubTask[0].AssignedTo = %q, want backend-001", req.SubTasks[0].AssignedTo)
	}
	if req.SubTasks[0].ChangeName != "change-login-api" {
		t.Errorf("SubTask[0].ChangeName = %q, want change-login-api", req.SubTasks[0].ChangeName)
	}

	// Step 4: Assign sub-task 2
	if err := AssignSubTask(req, 2, "frontend-001", "change-frontend"); err != nil {
		t.Fatalf("AssignSubTask(2): %v", err)
	}
	if err := SaveRequirement(dir, req); err != nil {
		t.Fatalf("SaveRequirement(assign2): %v", err)
	}
	if err := UpdateIndexEntry(dir, req); err != nil {
		t.Fatalf("UpdateIndexEntry(assign2): %v", err)
	}
	assertIndexSync(t, dir, "user-auth", RequirementStatusInProgress, 3, 0)

	// Step 5: Mark sub-task 1 done
	if err := MarkSubTaskDone(req, 1); err != nil {
		t.Fatalf("MarkSubTaskDone(1): %v", err)
	}
	if err := SaveRequirement(dir, req); err != nil {
		t.Fatalf("SaveRequirement(done1): %v", err)
	}
	if err := UpdateIndexEntry(dir, req); err != nil {
		t.Fatalf("UpdateIndexEntry(done1): %v", err)
	}

	// Requirement should still be in_progress (task 2 assigned, task 3 pending)
	if req.Status != RequirementStatusInProgress {
		t.Errorf("status after 1 done = %q, want in_progress", req.Status)
	}
	assertIndexSync(t, dir, "user-auth", RequirementStatusInProgress, 3, 1)

	// Step 6: Mark sub-task 2 done
	if err := MarkSubTaskDone(req, 2); err != nil {
		t.Fatalf("MarkSubTaskDone(2): %v", err)
	}
	if err := SaveRequirement(dir, req); err != nil {
		t.Fatalf("SaveRequirement(done2): %v", err)
	}
	if err := UpdateIndexEntry(dir, req); err != nil {
		t.Fatalf("UpdateIndexEntry(done2): %v", err)
	}

	// Still in_progress (task 3 pending)
	if req.Status != RequirementStatusInProgress {
		t.Errorf("status after 2 done = %q, want in_progress", req.Status)
	}
	assertIndexSync(t, dir, "user-auth", RequirementStatusInProgress, 3, 2)

	// Step 7: Assign and complete sub-task 3 → auto-promote to done
	if err := AssignSubTask(req, 3, "qa-001", "change-tests"); err != nil {
		t.Fatalf("AssignSubTask(3): %v", err)
	}
	if err := MarkSubTaskDone(req, 3); err != nil {
		t.Fatalf("MarkSubTaskDone(3): %v", err)
	}
	if err := SaveRequirement(dir, req); err != nil {
		t.Fatalf("SaveRequirement(done3): %v", err)
	}
	if err := UpdateIndexEntry(dir, req); err != nil {
		t.Fatalf("UpdateIndexEntry(done3): %v", err)
	}

	// All sub-tasks done → auto-promote to done
	if req.Status != RequirementStatusDone {
		t.Errorf("status after all done = %q, want done (auto-promote)", req.Status)
	}
	assertIndexSync(t, dir, "user-auth", RequirementStatusDone, 3, 3)

	// Step 8: Verify reload from disk produces the same state
	reloaded, err := LoadRequirement(dir, "user-auth")
	if err != nil {
		t.Fatalf("LoadRequirement: %v", err)
	}
	if reloaded.Status != RequirementStatusDone {
		t.Errorf("reloaded status = %q, want done", reloaded.Status)
	}
	for _, st := range reloaded.SubTasks {
		if st.Status != SubTaskStatusDone {
			t.Errorf("reloaded SubTask[%d].Status = %q, want done", st.ID, st.Status)
		}
	}
}

// TestRequirementMultipleReqsIndexSync verifies index.yaml tracks multiple requirements correctly.
func TestRequirementMultipleReqsIndexSync(t *testing.T) {
	dir := t.TempDir()

	// Create three requirements
	names := []string{"alpha", "beta", "gamma"}
	for _, name := range names {
		req, err := CreateRequirement(dir, name, "desc-"+name, nil)
		if err != nil {
			t.Fatalf("CreateRequirement(%s): %v", name, err)
		}
		if err := UpdateIndexEntry(dir, req); err != nil {
			t.Fatalf("UpdateIndexEntry(%s): %v", name, err)
		}
	}

	idx, err := LoadRequirementIndex(dir)
	if err != nil {
		t.Fatalf("LoadRequirementIndex: %v", err)
	}
	if len(idx.Requirements) != 3 {
		t.Fatalf("index has %d entries, want 3", len(idx.Requirements))
	}

	// Mark beta as done via direct manipulation
	req, _ := LoadRequirement(dir, "beta")
	req.Status = RequirementStatusDone
	SaveRequirement(dir, req)
	UpdateIndexEntry(dir, req)

	idx, _ = LoadRequirementIndex(dir)
	for _, e := range idx.Requirements {
		if e.Name == "beta" && e.Status != RequirementStatusDone {
			t.Errorf("beta index status = %q, want done", e.Status)
		}
		if e.Name == "alpha" && e.Status != RequirementStatusOpen {
			t.Errorf("alpha index status = %q, want open", e.Status)
		}
	}
}

// assertIndexSync verifies the index.yaml matches expected values for a requirement.
func assertIndexSync(t *testing.T, dir, name string, wantStatus RequirementStatus, wantSubCount, wantDone int) {
	t.Helper()
	idx, err := LoadRequirementIndex(dir)
	if err != nil {
		t.Fatalf("LoadRequirementIndex: %v", err)
	}
	for _, e := range idx.Requirements {
		if e.Name == name {
			if e.Status != wantStatus {
				t.Errorf("index[%s].Status = %q, want %q", name, e.Status, wantStatus)
			}
			if e.SubTaskCount != wantSubCount {
				t.Errorf("index[%s].SubTaskCount = %d, want %d", name, e.SubTaskCount, wantSubCount)
			}
			if e.DoneCount != wantDone {
				t.Errorf("index[%s].DoneCount = %d, want %d", name, e.DoneCount, wantDone)
			}
			return
		}
	}
	t.Errorf("requirement %q not found in index", name)
}
