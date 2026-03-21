package internal

import "fmt"

func ValidatePlanningReferences(root string, record *PlanningRecord) []string {
	if record == nil {
		return []string{"planning record is nil"}
	}
	var issues []string
	for _, id := range record.RoadmapIDs {
		issues = append(issues, validatePlanningRef(root, id, PlanningKindRoadmap)...)
	}
	for _, id := range record.MilestoneIDs {
		issues = append(issues, validatePlanningRef(root, id, PlanningKindMilestone)...)
	}
	for _, id := range record.PhaseIDs {
		issues = append(issues, validatePlanningRef(root, id, PlanningKindPhase)...)
	}
	for _, id := range record.TaskIDs {
		if _, _, err := LoadTaskRecord(root, id); err != nil {
			issues = append(issues, fmt.Sprintf("missing task reference: %s", id))
		}
	}
	return issues
}

func validatePlanningRef(root, id string, want PlanningKind) []string {
	record, err := LoadPlanningRecord(root, id)
	if err != nil {
		return []string{fmt.Sprintf("missing %s reference: %s", want, id)}
	}
	if record.Kind != want {
		return []string{fmt.Sprintf("reference kind mismatch: %s is %s, expected %s", id, record.Kind, want)}
	}
	return nil
}
