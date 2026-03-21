package internal

import (
	"os"
	"testing"
	"time"
)

func TestCreateAndLoadPlanningRecord(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	record, err := CreatePlanningRecord(root, PlanningKindRoadmap, "Planning MVP", now)
	if err != nil {
		t.Fatalf("CreatePlanningRecord: %v", err)
	}
	if record.Kind != PlanningKindRoadmap {
		t.Fatalf("Kind = %s", record.Kind)
	}
	if record.Lifecycle != PlanningLifecycleActive {
		t.Fatalf("Lifecycle = %s", record.Lifecycle)
	}
	if _, err := os.Stat(PlanningYAMLPath(root, record.Kind, record.ID, PlanningLifecycleActive)); err != nil {
		t.Fatalf("planning yaml missing: %v", err)
	}
	loaded, err := LoadPlanningRecord(root, record.ID)
	if err != nil {
		t.Fatalf("LoadPlanningRecord: %v", err)
	}
	if loaded.ID != record.ID || loaded.Path != record.Path {
		t.Fatalf("loaded = %#v", loaded)
	}
}

func TestMovePlanningRecord(t *testing.T) {
	root := t.TempDir()
	record, err := CreatePlanningRecord(root, PlanningKindPhase, "Phase One", time.Now().UTC())
	if err != nil {
		t.Fatalf("CreatePlanningRecord: %v", err)
	}
	moved, err := MovePlanningRecord(root, record.ID, PlanningLifecycleDeprecated, "replaced", time.Now().UTC())
	if err != nil {
		t.Fatalf("MovePlanningRecord: %v", err)
	}
	if moved.Lifecycle != PlanningLifecycleDeprecated {
		t.Fatalf("Lifecycle = %s", moved.Lifecycle)
	}
	if moved.DeprecatedReason != "replaced" {
		t.Fatalf("DeprecatedReason = %q", moved.DeprecatedReason)
	}
	if _, err := os.Stat(PlanningYAMLPath(root, moved.Kind, moved.ID, PlanningLifecycleDeprecated)); err != nil {
		t.Fatalf("deprecated planning yaml missing: %v", err)
	}
}
