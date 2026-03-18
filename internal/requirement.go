// internal/requirement.go
package internal

// RequirementStatus represents the lifecycle status of a Requirement.
type RequirementStatus string

const (
	RequirementStatusOpen       RequirementStatus = "open"
	RequirementStatusInProgress RequirementStatus = "in_progress"
	RequirementStatusDone       RequirementStatus = "done"
)

// SubTaskStatus represents the lifecycle status of a SubTask.
type SubTaskStatus string

const (
	SubTaskStatusPending  SubTaskStatus = "pending"
	SubTaskStatusAssigned SubTaskStatus = "assigned"
	SubTaskStatusDone     SubTaskStatus = "done"
	SubTaskStatusSkipped  SubTaskStatus = "skipped"
)

// SubTask represents a unit of work within a Requirement.
type SubTask struct {
	ID         int           `yaml:"id"`
	Title      string        `yaml:"title"`
	AssignedTo string        `yaml:"assigned_to,omitempty"`
	Status     SubTaskStatus `yaml:"status"`
	ChangeName string        `yaml:"change_name,omitempty"`
}

// Requirement represents a named requirement with sub-tasks.
type Requirement struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Status      RequirementStatus `yaml:"status"`
	CreatedAt   string            `yaml:"created_at"`
	SubTasks    []SubTask         `yaml:"sub_tasks,omitempty"`
}

// RequirementIndexEntry is a summary entry in the requirement index.
type RequirementIndexEntry struct {
	Name         string            `yaml:"name"`
	Status       RequirementStatus `yaml:"status"`
	SubTaskCount int               `yaml:"sub_task_count"`
	DoneCount    int               `yaml:"done_count"`
}

// RequirementIndex is the top-level index of all requirements.
type RequirementIndex struct {
	Requirements []RequirementIndexEntry `yaml:"requirements"`
}

// ValidRequirementStatus returns true if the status is valid.
func ValidRequirementStatus(s RequirementStatus) bool {
	switch s {
	case RequirementStatusOpen, RequirementStatusInProgress, RequirementStatusDone:
		return true
	default:
		return false
	}
}

// ValidSubTaskStatus returns true if the status is valid.
func ValidSubTaskStatus(s SubTaskStatus) bool {
	switch s {
	case SubTaskStatusPending, SubTaskStatusAssigned, SubTaskStatusDone, SubTaskStatusSkipped:
		return true
	default:
		return false
	}
}
