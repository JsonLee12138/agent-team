package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func stubProjectCommandsGenerator(t *testing.T, content string) {
	t.Helper()
	restore := SetProjectCommandsGenerator(func(root string, scan *BuildScriptScan) (string, error) {
		return content, nil
	})
	t.Cleanup(restore)
}

func TestDetectInstalledProviders(t *testing.T) {
	// This test just verifies the function runs without error.
	// Results depend on the host environment.
	providers := DetectInstalledProviders()
	for _, p := range providers {
		if p.Name == "" {
			t.Error("provider name should not be empty")
		}
		if p.Path == "" {
			t.Error("provider path should not be empty")
		}
	}
}

func TestHashLocalRoleDir(t *testing.T) {
	dir := t.TempDir()

	// Create some files
	os.MkdirAll(filepath.Join(dir, "references"), 0755)
	os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# test\n"), 0644)
	os.WriteFile(filepath.Join(dir, "system.md"), []byte("system prompt\n"), 0644)
	os.WriteFile(filepath.Join(dir, "references", "role.yaml"), []byte("name: test\n"), 0644)

	hash1, err := HashLocalRoleDir(dir)
	if err != nil {
		t.Fatalf("HashLocalRoleDir: %v", err)
	}
	if hash1 == "" {
		t.Error("hash should not be empty")
	}

	// Same content should produce same hash
	hash2, err := HashLocalRoleDir(dir)
	if err != nil {
		t.Fatalf("HashLocalRoleDir second call: %v", err)
	}
	if hash1 != hash2 {
		t.Error("same content should produce same hash")
	}

	// Modified content should produce different hash
	os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# modified\n"), 0644)
	hash3, err := HashLocalRoleDir(dir)
	if err != nil {
		t.Fatalf("HashLocalRoleDir after modify: %v", err)
	}
	if hash1 == hash3 {
		t.Error("different content should produce different hash")
	}
}

func TestHashLocalRoleDirEmpty(t *testing.T) {
	dir := t.TempDir()
	hash, err := HashLocalRoleDir(dir)
	if err != nil {
		t.Fatalf("HashLocalRoleDir empty dir: %v", err)
	}
	// Empty dir should still produce a valid hash
	if hash == "" {
		t.Error("empty dir should produce a hash")
	}
}

func TestInstallPluginRoleToGlobal_NewInstall(t *testing.T) {
	srcDir := t.TempDir()
	globalDir := t.TempDir()

	// Create source role
	os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("# role\n"), 0644)
	os.MkdirAll(filepath.Join(srcDir, "references"), 0755)
	os.WriteFile(filepath.Join(srcDir, "references", "role.yaml"), []byte("name: test-role\n"), 0644)

	hash, _ := HashLocalRoleDir(srcDir)
	candidate := PluginRoleCandidate{
		Name:    "test-role",
		Path:    srcDir,
		DirHash: hash,
	}

	action, err := InstallPluginRoleToGlobal(candidate, globalDir)
	if err != nil {
		t.Fatalf("InstallPluginRoleToGlobal: %v", err)
	}
	if action != InstallActionInstalled {
		t.Errorf("expected installed, got %s", action)
	}

	// Verify files exist
	dest := filepath.Join(globalDir, "test-role", "SKILL.md")
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Error("role files should be copied to global dir")
	}
}

func TestInstallPluginRoleToGlobal_Skipped(t *testing.T) {
	srcDir := t.TempDir()
	globalDir := t.TempDir()

	// Create source role
	os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("# role\n"), 0644)
	hash, _ := HashLocalRoleDir(srcDir)
	candidate := PluginRoleCandidate{
		Name:    "test-role",
		Path:    srcDir,
		DirHash: hash,
	}

	// Install once
	action, err := InstallPluginRoleToGlobal(candidate, globalDir)
	if err != nil {
		t.Fatalf("first install: %v", err)
	}
	if action != InstallActionInstalled {
		t.Fatalf("expected installed, got %s", action)
	}

	// Install again — should skip
	action, err = InstallPluginRoleToGlobal(candidate, globalDir)
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	if action != InstallActionSkipped {
		t.Errorf("expected skipped, got %s", action)
	}
}

func TestInstallPluginRoleToGlobal_Updated(t *testing.T) {
	srcDir := t.TempDir()
	globalDir := t.TempDir()

	// Create source role
	os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("# role v1\n"), 0644)
	hash1, _ := HashLocalRoleDir(srcDir)
	candidate := PluginRoleCandidate{
		Name:    "test-role",
		Path:    srcDir,
		DirHash: hash1,
	}

	// Install
	_, err := InstallPluginRoleToGlobal(candidate, globalDir)
	if err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Modify source
	os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("# role v2\n"), 0644)
	hash2, _ := HashLocalRoleDir(srcDir)
	candidate.DirHash = hash2

	// Install again — should update
	action, err := InstallPluginRoleToGlobal(candidate, globalDir)
	if err != nil {
		t.Fatalf("update install: %v", err)
	}
	if action != InstallActionUpdated {
		t.Errorf("expected updated, got %s", action)
	}

	// Verify updated content
	content, _ := os.ReadFile(filepath.Join(globalDir, "test-role", "SKILL.md"))
	if string(content) != "# role v2\n" {
		t.Errorf("role should be updated, got %q", string(content))
	}
}

func TestInitProject(t *testing.T) {
	dir := t.TempDir()

	err := InitProject(dir)
	if err != nil {
		t.Fatalf("InitProject: %v", err)
	}

	teamsDir := filepath.Join(dir, ".agent-team", "teams")
	if _, err := os.Stat(teamsDir); os.IsNotExist(err) {
		t.Error(".agent-team/teams/ should be created")
	}

	gitkeep := filepath.Join(teamsDir, ".gitkeep")
	if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
		t.Error(".gitkeep should be created")
	}

	// Running again should not error
	err = InitProject(dir)
	if err != nil {
		t.Fatalf("InitProject second run: %v", err)
	}
}

func TestScanPluginRoles_NoEnv(t *testing.T) {
	t.Setenv("CLAUDE_PLUGIN_ROOT", "")
	candidates := ScanPluginRoles()
	if candidates != nil {
		t.Error("should return nil when CLAUDE_PLUGIN_ROOT is empty")
	}
}

func TestScanPluginRoles_WithRoles(t *testing.T) {
	pluginRoot := t.TempDir()

	// Create a role with references/role.yaml
	roleDir := filepath.Join(pluginRoot, "skills", "my-role")
	os.MkdirAll(filepath.Join(roleDir, "references"), 0755)
	os.WriteFile(filepath.Join(roleDir, "references", "role.yaml"), []byte("name: my-role\n"), 0644)
	os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# my-role\n"), 0644)

	// Create a non-role directory (no role.yaml)
	nonRole := filepath.Join(pluginRoot, "skills", "not-a-role")
	os.MkdirAll(nonRole, 0755)
	os.WriteFile(filepath.Join(nonRole, "SKILL.md"), []byte("# not a role\n"), 0644)

	t.Setenv("CLAUDE_PLUGIN_ROOT", pluginRoot)

	candidates := ScanPluginRoles()
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0].Name != "my-role" {
		t.Errorf("expected 'my-role', got %q", candidates[0].Name)
	}
	if candidates[0].DirHash == "" {
		t.Error("candidate should have a hash")
	}
}

func TestFormatProviderList(t *testing.T) {
	providers := []DetectedProvider{
		{Name: "claude", Path: "/usr/bin/claude"},
		{Name: "gemini", Path: "/usr/bin/gemini"},
	}
	result := FormatProviderList(providers)
	if result != "claude, gemini" {
		t.Errorf("expected 'claude, gemini', got %q", result)
	}
}

func TestDetectProjectBuildScripts(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte(".PHONY: build test lint\nbuild:\n\tgo build ./...\n\ntest:\n\tgo test ./...\n\nlint:\n\tgo vet ./...\n"), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/project\n\ngo 1.24.2\n"), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "web"), 0755); err != nil {
		t.Fatalf("mkdir web: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "web", "package.json"), []byte(`{"name":"web","scripts":{"build":"vite build","lint":"eslint .","test":"vitest run"}}`), 0644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}

	scan, err := DetectProjectBuildScripts(dir)
	if err != nil {
		t.Fatalf("DetectProjectBuildScripts: %v", err)
	}
	if scan.Hash == "" {
		t.Fatal("hash should not be empty")
	}
	if len(scan.Files) != 3 {
		t.Fatalf("Files = %v, want 3 entries", scan.Files)
	}
	if !strings.Contains(strings.Join(scan.MakeTargets, ","), "build") {
		t.Fatalf("MakeTargets = %v, want build target", scan.MakeTargets)
	}
	if len(scan.GoModules) != 1 || scan.GoModules[0].Module != "example.com/project" {
		t.Fatalf("GoModules = %+v, want example.com/project", scan.GoModules)
	}
	if len(scan.NodePackages) != 1 || scan.NodePackages[0].Path != filepath.Join("web", "package.json") {
		t.Fatalf("NodePackages = %+v, want web/package.json", scan.NodePackages)
	}
	if len(scan.NodePackages[0].Scripts) != 3 {
		t.Fatalf("Scripts = %+v, want 3 scripts", scan.NodePackages[0].Scripts)
	}
}

func TestRebuildProjectCommands(t *testing.T) {
	dir := t.TempDir()
	stubProjectCommandsGenerator(t, "# Project Commands Rules\n\nGenerated for tests.\n")

	if err := os.MkdirAll(filepath.Join(dir, ".agent-team"), 0755); err != nil {
		t.Fatalf("mkdir .agent-team: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:\n\tgo build ./...\n\ntest:\n\tgo test ./...\n"), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/project\n\ngo 1.24.2\n"), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "role-hub"), 0755); err != nil {
		t.Fatalf("mkdir role-hub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "role-hub", "package.json"), []byte(`{"name":"role-hub","scripts":{"build":"remix build","test":"vitest run"}}`), 0644); err != nil {
		t.Fatalf("write role-hub/package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".build-hash"), []byte("legacy\n"), 0644); err != nil {
		t.Fatalf("write .build-hash: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".agent-team", "rules", "core"), 0755); err != nil {
		t.Fatalf("mkdir .agent-team/rules: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".agent-team", "rules", "build-verification.md"), []byte("legacy\n"), 0644); err != nil {
		t.Fatalf("write build-verification.md: %v", err)
	}

	scan, err := RebuildProjectCommands(dir)
	if err != nil {
		t.Fatalf("RebuildProjectCommands: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".agent-team", "rules", "project", projectCommandsFileName))
	if err != nil {
		t.Fatalf("read %s: %v", projectCommandsFileName, err)
	}
	content := string(data)
	for _, needle := range []string{
		"# Project Commands Rules",
		"Generated for tests.",
	} {
		if !strings.Contains(content, needle) {
			t.Fatalf("%s missing %q\n%s", projectCommandsFileName, needle, content)
		}
	}
	if len(scan.Files) != 3 {
		t.Fatalf("Files = %v, want 3 entries", scan.Files)
	}
	if _, err := os.Stat(filepath.Join(dir, ".build-hash")); !os.IsNotExist(err) {
		t.Fatalf(".build-hash should be removed, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agent-team", "rules", "build-verification.md")); !os.IsNotExist(err) {
		t.Fatalf("legacy build-verification.md should be removed, err=%v", err)
	}
}

func TestInitRulesDir(t *testing.T) {
	t.Run("creates all default rule files", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agent-team"), 0755)

		created, err := InitRulesDir(dir)
		if err != nil {
			t.Fatalf("InitRulesDir: %v", err)
		}
		wantCreated := len(defaultRuleFiles) + len(defaultCoreRuleFiles)
		if created != wantCreated {
			t.Errorf("created %d files, want %d", created, wantCreated)
		}

		// Check all files exist
		rulesDir := filepath.Join(dir, ".agent-team", "rules")
		if _, err := os.Stat(filepath.Join(rulesDir, "index.md")); os.IsNotExist(err) {
			t.Error("expected index.md to exist")
		}
		for name := range defaultCoreRuleFiles {
			fp := filepath.Join(rulesDir, "core", name)
			if _, err := os.Stat(fp); os.IsNotExist(err) {
				t.Errorf("expected core/%s to exist", name)
			}
		}
		if _, err := os.Stat(filepath.Join(rulesDir, "project", projectCommandsFileName)); !os.IsNotExist(err) {
			t.Error("project/project-commands.md should not be created by InitRulesDir")
		}
		if _, err := os.Stat(filepath.Join(dir, ".build-hash")); !os.IsNotExist(err) {
			t.Error(".build-hash should not exist after InitRulesDir")
		}
	})

	t.Run("idempotent - does not overwrite existing files", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agent-team", "rules", "core"), 0755)

		// Write a custom core rule file
		customContent := "# Custom rules\n"
		os.WriteFile(filepath.Join(dir, ".agent-team", "rules", "core", "debugging.md"), []byte(customContent), 0644)

		created, err := InitRulesDir(dir)
		if err != nil {
			t.Fatalf("InitRulesDir: %v", err)
		}
		// Should create index.md plus all missing core files except the customized one
		wantCreated := len(defaultRuleFiles) + len(defaultCoreRuleFiles) - 1
		if created != wantCreated {
			t.Errorf("created %d files, want %d", created, wantCreated)
		}

		// Verify custom content is preserved
		data, _ := os.ReadFile(filepath.Join(dir, ".agent-team", "rules", "core", "debugging.md"))
		if string(data) != customContent {
			t.Errorf("custom content should be preserved, got %q", string(data))
		}
	})

	t.Run("second run creates zero files", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agent-team"), 0755)

		InitRulesDir(dir)
		created, err := InitRulesDir(dir)
		if err != nil {
			t.Fatalf("second InitRulesDir: %v", err)
		}
		if created != 0 {
			t.Errorf("second run should create 0 files, created %d", created)
		}
	})
}

func TestSyncRulesDir(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, ".agent-team", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("mkdir rules dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "index.md"), []byte("# stale\n"), 0644); err != nil {
		t.Fatalf("write stale index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "build-verification.md"), []byte("legacy\n"), 0644); err != nil {
		t.Fatalf("write legacy build-verification.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".build-hash"), []byte("legacy\n"), 0644); err != nil {
		t.Fatalf("write .build-hash: %v", err)
	}

	written, err := SyncRulesDir(dir)
	if err != nil {
		t.Fatalf("SyncRulesDir: %v", err)
	}
	wantWritten := len(defaultRuleFiles) + len(defaultCoreRuleFiles)
	if written != wantWritten {
		t.Fatalf("written %d files, want %d", written, wantWritten)
	}

	indexData, err := os.ReadFile(filepath.Join(rulesDir, "index.md"))
	if err != nil {
		t.Fatalf("read index.md: %v", err)
	}
	for _, needle := range []string{"core/debugging.md", "core/agent-team-commands.md", "core/merge-workflow.md", "project/", "worktree.md"} {
		if !strings.Contains(string(indexData), needle) {
			t.Fatalf("index.md should reference %s, got:\n%s", needle, string(indexData))
		}
	}
	if _, err := os.Stat(filepath.Join(rulesDir, "build-verification.md")); !os.IsNotExist(err) {
		t.Fatalf("legacy build-verification.md should be removed, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".build-hash")); !os.IsNotExist(err) {
		t.Fatalf(".build-hash should be removed, err=%v", err)
	}
}

func TestInitProviderFiles(t *testing.T) {
	t.Run("creates new provider files", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agent-team", "rules", "core"), 0755)

		err := InitProviderFiles(dir)
		if err != nil {
			t.Fatalf("InitProviderFiles: %v", err)
		}

		for _, name := range []string{"CLAUDE.md", "AGENTS.md", "GEMINI.md"} {
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatalf("read %s: %v", name, err)
			}
			content := string(data)
			if !strings.Contains(content, "<!-- AGENT_TEAM:START -->") {
				t.Errorf("%s should contain AGENT_TEAM start marker", name)
			}
			if !strings.Contains(content, "Rules Reference") {
				t.Errorf("%s should contain Rules Reference", name)
			}
			if !strings.Contains(content, ".agent-team/rules/index.md") {
				t.Errorf("%s should reference rules/index.md", name)
			}
			if strings.Contains(content, "MUST call `/compact`") {
				t.Errorf("%s should not require /compact", name)
			}
			for _, needle := range []string{"context-cleanup", "index-first recovery", ".agent-team/rules/core/context-management.md", ".agent-team/rules/project/", ".agent-team/rules/core/agent-team-commands.md", ".agent-team/rules/core/merge-workflow.md", ".agent-team/rules/core/worktree.md"} {
				if !strings.Contains(content, needle) {
					t.Errorf("%s should reference %s", name, needle)
				}
			}
		}

		settingsData, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.local.json"))
		if err != nil {
			t.Fatalf("read settings.local.json: %v", err)
		}
		content := string(settingsData)
		if !strings.Contains(content, "SessionStart") {
			t.Fatal("settings.local.json should contain SessionStart hook")
		}
		if !strings.Contains(content, "./scripts/session-start-record-main-pane.sh") {
			t.Fatal("settings.local.json should contain record-main-pane script")
		}
	})

	t.Run("preserves user content when updating", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agent-team", "rules", "core"), 0755)

		// Write custom CLAUDE.md with user content
		userContent := "# My Custom Project\n\nThis is my project.\n"
		os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(userContent), 0644)

		err := InitProviderFiles(dir)
		if err != nil {
			t.Fatalf("InitProviderFiles: %v", err)
		}

		data, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
		content := string(data)
		if !strings.Contains(content, "My Custom Project") {
			t.Error("user content should be preserved")
		}
		if !strings.Contains(content, "<!-- AGENT_TEAM:START -->") {
			t.Error("tag section should be injected")
		}
	})

	t.Run("updates only tag section on re-run", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agent-team", "rules", "core"), 0755)

		// First run
		InitProviderFiles(dir)

		// Add user content after the tag section
		fp := filepath.Join(dir, "CLAUDE.md")
		existing, _ := os.ReadFile(fp)
		os.WriteFile(fp, append(existing, []byte("\n## My Custom Section\n\nUser notes.\n")...), 0644)

		// Second run
		err := InitProviderFiles(dir)
		if err != nil {
			t.Fatalf("InitProviderFiles second run: %v", err)
		}

		data, _ := os.ReadFile(fp)
		content := string(data)
		if !strings.Contains(content, "My Custom Section") {
			t.Error("user-added section should be preserved")
		}
		if strings.Count(content, "<!-- AGENT_TEAM:START -->") != 1 {
			t.Error("should have exactly one start marker")
		}
	})

	t.Run("merges settings.local hooks without clobbering existing config", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".agent-team", "rules", "core"), 0755)
		if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0755); err != nil {
			t.Fatalf("mkdir .claude: %v", err)
		}
		original := `{
  "permissions": {
    "allow": ["Bash(git:*)"]
  },
  "hooks": {
    "SessionStart": [
      {
        "matcher": "*",
        "hooks": [
          {
            "name": "existing-hook",
            "type": "command",
            "command": "./scripts/existing.sh",
            "timeout": 5000
          }
        ]
      }
    ]
  }
}
`
		settingsPath := filepath.Join(dir, ".claude", "settings.local.json")
		if err := os.WriteFile(settingsPath, []byte(original), 0644); err != nil {
			t.Fatalf("write settings.local.json: %v", err)
		}

		if err := InitProviderFiles(dir); err != nil {
			t.Fatalf("InitProviderFiles: %v", err)
		}

		data, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("read settings.local.json: %v", err)
		}
		content := string(data)
		if !strings.Contains(content, "Bash(git:*)") {
			t.Fatal("existing permissions should be preserved")
		}
		if !strings.Contains(content, "./scripts/existing.sh") {
			t.Fatal("existing SessionStart hook should be preserved")
		}
		if !strings.Contains(content, "./scripts/session-start-record-main-pane.sh") {
			t.Fatal("record-main-pane hook should be added")
		}
		if strings.Count(content, "./scripts/session-start-record-main-pane.sh") != 1 {
			t.Fatal("record-main-pane hook should not be duplicated")
		}
	})
}
