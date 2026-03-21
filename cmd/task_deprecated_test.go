package cmd

import (
	"os"
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunTaskDeprecatedMovesTask(t *testing.T) {
	app, dir := initTestApp(t)
	record, err := internal.CreateTaskPackage(dir, "Deprecated Task", "backend", "", time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("CreateTaskPackage: %v", err)
	}
	if err := app.RunTaskDeprecated(record.TaskID); err != nil {
		t.Fatalf("RunTaskDeprecated: %v", err)
	}
	loaded, location, err := internal.LoadTaskRecord(dir, record.TaskID)
	if err != nil {
		t.Fatalf("LoadTaskRecord: %v", err)
	}
	if location != internal.TaskRecordLocationDeprecated {
		t.Fatalf("location = %s", location)
	}
	if loaded.Status != internal.TaskStatusDeprecated {
		t.Fatalf("status = %s", loaded.Status)
	}
	if _, err := os.Stat(internal.TaskDeprecatedVerificationPath(dir, record.TaskID)); err != nil {
		t.Fatalf("deprecated verification missing: %v", err)
	}
}
