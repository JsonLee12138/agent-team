package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/JsonLee12138/agent-team/internal"
)

func TestRunPlanningShowDisplaysReferenceChecks(t *testing.T) {
	app, dir := initTestApp(t)
	record, err := internal.CreatePlanningRecord(dir, internal.PlanningKindPhase, "Phase A", time.Now().UTC())
	if err != nil {
		t.Fatalf("CreatePlanningRecord: %v", err)
	}
	out := captureStdout(t, func() {
		if err := app.RunPlanningShow(record.ID); err != nil {
			t.Fatalf("RunPlanningShow: %v", err)
		}
	})
	for _, needle := range []string{"ID:", "Kind: phase", "Lifecycle: planning", "Reference Checks: ok"} {
		if !strings.Contains(out, needle) {
			t.Fatalf("output missing %q:\n%s", needle, out)
		}
	}
}
