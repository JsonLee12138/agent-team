package internal

import "fmt"

// PlanningKind identifies the planning artifact type.
type PlanningKind string

const (
	PlanningKindRoadmap   PlanningKind = "roadmap"
	PlanningKindMilestone PlanningKind = "milestone"
	PlanningKindPhase     PlanningKind = "phase"
)

// PlanningLifecycle identifies where the planning artifact is stored.
type PlanningLifecycle string

const (
	PlanningLifecycleActive     PlanningLifecycle = "planning"
	PlanningLifecycleArchived   PlanningLifecycle = "archived"
	PlanningLifecycleDeprecated PlanningLifecycle = "deprecated"
)

// PlanningRecord is the structured record stored in roadmap/milestone/phase yaml files.
type PlanningRecord struct {
	ID               string            `yaml:"id"`
	Kind             PlanningKind      `yaml:"kind"`
	Title            string            `yaml:"title"`
	Status           string            `yaml:"status,omitempty"`
	Goal             string            `yaml:"goal,omitempty"`
	Lifecycle        PlanningLifecycle `yaml:"lifecycle"`
	Path             string            `yaml:"path"`
	RoadmapIDs       []string          `yaml:"roadmap_ids,omitempty"`
	MilestoneIDs     []string          `yaml:"milestone_ids,omitempty"`
	PhaseIDs         []string          `yaml:"phase_ids,omitempty"`
	TaskIDs          []string          `yaml:"task_ids,omitempty"`
	CreatedAt        string            `yaml:"created_at"`
	UpdatedAt        string            `yaml:"updated_at"`
	ArchivedAt       string            `yaml:"archived_at,omitempty"`
	DeprecatedAt     string            `yaml:"deprecated_at,omitempty"`
	DeprecatedReason string            `yaml:"deprecated_reason,omitempty"`
	ReplacedBy       string            `yaml:"replaced_by,omitempty"`
}

func ValidPlanningKind(kind PlanningKind) bool {
	switch kind {
	case PlanningKindRoadmap, PlanningKindMilestone, PlanningKindPhase:
		return true
	default:
		return false
	}
}

func ParsePlanningKind(raw string) (PlanningKind, error) {
	kind := PlanningKind(raw)
	if !ValidPlanningKind(kind) {
		return "", fmt.Errorf("invalid planning kind: %s", raw)
	}
	return kind, nil
}

func ValidPlanningLifecycle(lifecycle PlanningLifecycle) bool {
	switch lifecycle {
	case PlanningLifecycleActive, PlanningLifecycleArchived, PlanningLifecycleDeprecated:
		return true
	default:
		return false
	}
}

func ParsePlanningLifecycle(raw string) (PlanningLifecycle, error) {
	lifecycle := PlanningLifecycle(raw)
	if !ValidPlanningLifecycle(lifecycle) {
		return "", fmt.Errorf("invalid planning lifecycle: %s", raw)
	}
	return lifecycle, nil
}

func planningDirName(kind PlanningKind) string {
	switch kind {
	case PlanningKindRoadmap:
		return "roadmaps"
	case PlanningKindMilestone:
		return "milestones"
	default:
		return "phases"
	}
}

func planningFileName(kind PlanningKind) string {
	return string(kind) + ".yaml"
}
