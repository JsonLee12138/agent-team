package cmd

import (
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunPlanningMoveArchivesArtifact(t *testing.T) {
	app, dir := initTestApp(t)
	record, err := internal.CreatePlanningRecord(dir, internal.PlanningKindMilestone, "Milestone A", time.Now().UTC())
	if err != nil {
		t.Fatalf("CreatePlanningRecord: %v", err)
	}
	if err := app.RunPlanningMove(record.ID, "archived", ""); err != nil {
		t.Fatalf("RunPlanningMove: %v", err)
	}
	loaded, err := internal.LoadPlanningRecord(dir, record.ID)
	if err != nil {
		t.Fatalf("LoadPlanningRecord: %v", err)
	}
	if loaded.Lifecycle != internal.PlanningLifecycleArchived {
		t.Fatalf("Lifecycle = %s", loaded.Lifecycle)
	}
}
