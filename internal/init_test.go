package internal

import (
	"os"
	"path/filepath"
	"testing"
)

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

	teamsDir := filepath.Join(dir, ".agents", "teams")
	if _, err := os.Stat(teamsDir); os.IsNotExist(err) {
		t.Error(".agents/teams/ should be created")
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
