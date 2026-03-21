package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunPlanningListPrintsArtifacts(t *testing.T) {
	app, dir := initTestApp(t)
	if _, err := internal.CreatePlanningRecord(dir, internal.PlanningKindRoadmap, "Roadmap A", time.Now().UTC()); err != nil {
		t.Fatalf("CreatePlanningRecord: %v", err)
	}
	out := captureStdout(t, func() {
		if err := app.RunPlanningList("", ""); err != nil {
			t.Fatalf("RunPlanningList: %v", err)
		}
	})
	for _, needle := range []string{"Kind", "Lifecycle", "roadmap", "Roadmap A"} {
		if !strings.Contains(out, needle) {
			t.Fatalf("output missing %q:\n%s", needle, out)
		}
	}
}
