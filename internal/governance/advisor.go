package governance

import (
	"fmt"
	"time"
)

func GenerateWorkflowPlan(input AdvisorInput) (*WorkflowPlan, error) {
	if input.PlanID == "" {
		return nil, fmt.Errorf("plan id is required")
	}
	if input.TaskPacket.TaskID == "" {
		return nil, fmt.Errorf("task id is required")
	}
	if input.TaskPacket.Owner == "" {
		return nil, fmt.Errorf("owner is required")
	}

	now := input.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	refs := dedupeStrings(append(append([]string{input.TaskPacket.TaskID}, input.TaskPacket.DeclaredReferences...), input.EvidenceRefs...))
	reasons := append([]string(nil), input.Reasons...)
	if len(reasons) == 0 {
		reasons = []string{"generated from text artifacts only"}
	}

	plan := NewWorkflowPlan(input.PlanID, input.TaskPacket.TaskID, input.TaskPacket.Owner, refs, reasons, now)
	plan.Status = WorkflowPlanStatusProposed
	return plan, nil
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
