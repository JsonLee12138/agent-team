package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunPlanningCreateCreatesArtifact(t *testing.T) {
	app, dir := initTestApp(t)
	if err := app.RunPlanningCreate("roadmap", "Platform Planning"); err != nil {
		t.Fatalf("RunPlanningCreate: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, ".agent-team", "planning", "roadmaps"))
	if err != nil {
		t.Fatalf("ReadDir roadmaps: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	record, err := internal.LoadPlanningRecord(dir, entries[0].Name())
	if err != nil {
		t.Fatalf("LoadPlanningRecord: %v", err)
	}
	if record.Kind != internal.PlanningKindRoadmap {
		t.Fatalf("Kind = %s", record.Kind)
	}
}
