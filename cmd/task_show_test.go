package cmd

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunTaskShowLoadsTaskRecord(t *testing.T) {
	app, dir := initTestApp(t)
	record, err := internal.CreateTaskPackage(dir, "Show Task", "backend", "", time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("CreateTaskPackage: %v", err)
	}
	if err := os.WriteFile(internal.TaskVerificationPath(dir, record.TaskID), []byte("# Verification\n\n## Result\n- partial\n"), 0644); err != nil {
		t.Fatalf("WriteFile verification: %v", err)
	}
	out := captureStdout(t, func() {
		if err := app.RunTaskShow(record.TaskID); err != nil {
			t.Fatalf("RunTaskShow: %v", err)
		}
	})
	for _, needle := range []string{"# Verification", "Verification Result: partial", "Archive Ready (default): yes", "Archive Ready (strict): no"} {
		if !strings.Contains(out, needle) {
			t.Fatalf("output should include %q, got:\n%s", needle, out)
		}
	}
}
