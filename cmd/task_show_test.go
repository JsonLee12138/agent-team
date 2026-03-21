package cmd

import (
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
	out := captureStdout(t, func() {
		if err := app.RunTaskShow(record.TaskID); err != nil {
			t.Fatalf("RunTaskShow: %v", err)
		}
	})
	if !strings.Contains(out, "# Verification") {
		t.Fatalf("output should include verification.md, got:\n%s", out)
	}
}
