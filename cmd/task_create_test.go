package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunTaskCreateCreatesTaskPackage(t *testing.T) {
	app, dir := initTestApp(t)
	roleDir := filepath.Join(dir, ".agent-team", "teams", "backend")
	if err := os.MkdirAll(roleDir, 0755); err != nil {
		t.Fatalf("mkdir role dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(roleDir, "SKILL.md"), []byte("# backend\n"), 0644); err != nil {
		t.Fatalf("write skill: %v", err)
	}
	if err := app.RunTaskCreate("Implement lifecycle", "backend", ""); err != nil {
		t.Fatalf("RunTaskCreate: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, ".agent-team", "task"))
	if err != nil {
		t.Fatalf("ReadDir task root: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
}
