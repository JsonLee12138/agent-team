package internal

import (
	"strings"
	"testing"
	"time"
)

func TestValidatePlanningReferencesReportsMissingRefs(t *testing.T) {
	root := t.TempDir()
	record, err := CreatePlanningRecord(root, PlanningKindMilestone, "Milestone", time.Now().UTC())
	if err != nil {
		t.Fatalf("CreatePlanningRecord: %v", err)
	}
	record.TaskIDs = []string{"missing-task"}
	record.PhaseIDs = []string{"missing-phase"}
	issues := ValidatePlanningReferences(root, record)
	joined := strings.Join(issues, "\n")
	for _, needle := range []string{"missing task reference: missing-task", "missing phase reference: missing-phase"} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("issues missing %q:\n%s", needle, joined)
		}
	}
}
