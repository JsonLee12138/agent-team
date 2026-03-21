package internal

// TaskStatus represents the lifecycle status of a task package.
type TaskStatus string

const (
	TaskStatusDraft     TaskStatus = "draft"
	TaskStatusAssigned  TaskStatus = "assigned"
	TaskStatusVerifying TaskStatus = "verifying"
	TaskStatusArchived  TaskStatus = "archived"
	TaskStatusDeprecated TaskStatus = "deprecated"
)

// TaskRecordLocation represents where the task package is stored.
type TaskRecordLocation string

const (
	TaskRecordLocationActive     TaskRecordLocation = "active"
	TaskRecordLocationArchived   TaskRecordLocation = "archived"
	TaskRecordLocationDeprecated TaskRecordLocation = "deprecated"
)

// TaskRecord is the structured record stored in task.yaml.
type TaskRecord struct {
	TaskID        string     `yaml:"task_id"`
	Title         string     `yaml:"title"`
	Role          string     `yaml:"role"`
	Status        TaskStatus `yaml:"status"`
	WorkerID      string     `yaml:"worker_id,omitempty"`
	TaskPath      string     `yaml:"task_path"`
	CreatedAt     string     `yaml:"created_at"`
	AssignedAt    string     `yaml:"assigned_at,omitempty"`
	VerifyingAt   string     `yaml:"verifying_at,omitempty"`
	ArchivedAt    string     `yaml:"archived_at,omitempty"`
	DeprecatedAt  string     `yaml:"deprecated_at,omitempty"`
	MergedSHA     string     `yaml:"merged_sha,omitempty"`
	LegacyDoneAt  string     `yaml:"done_at,omitempty"`
}

func ValidTaskStatus(status TaskStatus) bool {
	switch status {
	case TaskStatusDraft, TaskStatusAssigned, TaskStatusVerifying, TaskStatusArchived, TaskStatusDeprecated:
		return true
	default:
		return false
	}
}
