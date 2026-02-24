// internal/openspec_test.go
package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestEnsureOpenSpec_AlreadyInstalled(t *testing.T) {
	// If openspec is on PATH, EnsureOpenSpec should succeed without installing
	if _, err := exec.LookPath("openspec"); err != nil {
		t.Skip("openspec not installed, skipping")
	}
	err := EnsureOpenSpec()
	if err != nil {
		t.Fatalf("EnsureOpenSpec: %v", err)
	}
}

func TestOpenSpecInit(t *testing.T) {
	if _, err := exec.LookPath("openspec"); err != nil {
		t.Skip("openspec not installed, skipping")
	}

	dir := t.TempDir()
	// openspec init requires a git repo
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s (%v)", args, out, err)
		}
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	run("commit", "--allow-empty", "-m", "init")

	err := OpenSpecInit(dir)
	if err != nil {
		t.Fatalf("OpenSpecInit: %v", err)
	}

	// Verify openspec directory was created
	if _, err := os.Stat(filepath.Join(dir, "openspec")); os.IsNotExist(err) {
		t.Error("openspec/ directory not created")
	}
	if _, err := os.Stat(filepath.Join(dir, "openspec", "changes")); os.IsNotExist(err) {
		t.Error("openspec/changes/ directory not created")
	}
}

func TestCreateChange(t *testing.T) {
	dir := t.TempDir()
	// Create minimal openspec structure
	os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0755)

	changeName := "2026-02-24-fix-login"
	proposal := "# Proposal\n\nFix the login flow by adding JWT validation."

	changePath, err := CreateChange(dir, changeName, proposal)
	if err != nil {
		t.Fatalf("CreateChange: %v", err)
	}

	// Verify change directory
	if _, err := os.Stat(changePath); os.IsNotExist(err) {
		t.Error("change directory not created")
	}

	// Verify .openspec.yaml
	metaPath := filepath.Join(changePath, ".openspec.yaml")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error(".openspec.yaml not created")
	}

	// Verify proposal.md
	proposalPath := filepath.Join(changePath, "proposal.md")
	data, err := os.ReadFile(proposalPath)
	if err != nil {
		t.Fatalf("read proposal.md: %v", err)
	}
	if string(data) != proposal {
		t.Errorf("proposal content = %q, want %q", string(data), proposal)
	}
}

func TestCreateChangeEmptyProposal(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0755)

	changePath, err := CreateChange(dir, "2026-02-24-test", "")
	if err != nil {
		t.Fatalf("CreateChange: %v", err)
	}

	// proposal.md should exist but be empty
	data, _ := os.ReadFile(filepath.Join(changePath, "proposal.md"))
	if len(data) != 0 {
		t.Errorf("expected empty proposal, got %q", string(data))
	}
}

func TestParseOpenSpecStatus(t *testing.T) {
	// Test parsing the JSON output from openspec status
	jsonData := `{"changes":[{"name":"fix-login","artifacts":{"proposal":{"status":"done"},"specs":{"status":"ready"},"design":{"status":"blocked"},"tasks":{"status":"blocked"}}},{"name":"add-auth","artifacts":{"proposal":{"status":"done"},"specs":{"status":"done"},"design":{"status":"done"},"tasks":{"status":"done"}}}]}`

	result, err := ParseOpenSpecStatus([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseOpenSpecStatus: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(result))
	}
	if result[0].Name != "fix-login" {
		t.Errorf("change[0].Name = %q, want fix-login", result[0].Name)
	}
	if result[0].Phase != "planning" {
		t.Errorf("change[0].Phase = %q, want planning", result[0].Phase)
	}
	if result[1].Phase != "ready" {
		t.Errorf("change[1].Phase = %q, want ready", result[1].Phase)
	}
}
