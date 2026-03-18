package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRulesSyncCmdHasRebuildFlag(t *testing.T) {
	cmd := newRulesSyncCmd()
	if flag := cmd.Flags().Lookup("rebuild"); flag == nil {
		t.Fatal("sync command should expose --rebuild")
	}
}

func TestRulesSyncRebuildRefreshesBuildVerification(t *testing.T) {
	_, dir := initTestApp(t)

	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:\n\tgo build ./...\n"), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}
	if _, err := internal.RebuildBuildRules(dir); err != nil {
		t.Fatalf("RebuildBuildRules: %v", err)
	}

	oldHash, err := internal.ReadBuildHash(dir)
	if err != nil {
		t.Fatalf("ReadBuildHash: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:\n\tgo build ./...\n\ntest:\n\tgo test ./...\n"), 0644); err != nil {
		t.Fatalf("update Makefile: %v", err)
	}

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer os.Chdir(origWd)

	root := NewRootCmd()
	RegisterCommands(root)
	root.SetArgs([]string{"rules", "sync", "--rebuild"})
	if err := root.Execute(); err != nil {
		t.Fatalf("rules sync --rebuild: %v", err)
	}

	newHash, err := internal.ReadBuildHash(dir)
	if err != nil {
		t.Fatalf("ReadBuildHash after rebuild: %v", err)
	}
	if newHash == oldHash {
		t.Fatalf("hash should change after Makefile update, old=%q new=%q", oldHash, newHash)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".agents", "rules", "build-verification.md"))
	if err != nil {
		t.Fatalf("read build-verification.md: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "`make test`") {
		t.Fatalf("build-verification.md = %q, want rebuilt make test target", content)
	}
	if !strings.Contains(content, "Current build-script hash: `"+newHash+"`") {
		t.Fatalf("build-verification.md should contain refreshed hash %q", newHash)
	}
}
