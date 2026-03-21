package internal

// TaskStatus represents the lifecycle status of a task package.
type TaskStatus string

const (
	TaskStatusDraft    TaskStatus = "draft"
	TaskStatusAssigned TaskStatus = "assigned"
	TaskStatusDone     TaskStatus = "done"
	TaskStatusArchived TaskStatus = "archived"
)

// TaskRecord is the structured record stored in task.yaml.
type TaskRecord struct {
	TaskID     string     `yaml:"task_id"`
	Title      string     `yaml:"title"`
	Role       string     `yaml:"role"`
	Status     TaskStatus `yaml:"status"`
	WorkerID   string     `yaml:"worker_id,omitempty"`
	TaskPath   string     `yaml:"task_path"`
	CreatedAt  string     `yaml:"created_at"`
	AssignedAt string     `yaml:"assigned_at,omitempty"`
	DoneAt     string     `yaml:"done_at,omitempty"`
	ArchivedAt string     `yaml:"archived_at,omitempty"`
	MergedSHA  string     `yaml:"merged_sha,omitempty"`
}

func ValidTaskStatus(status TaskStatus) bool {
	switch status {
	case TaskStatusDraft, TaskStatusAssigned, TaskStatusDone, TaskStatusArchived:
		return true
	default:
		return false
	}
}
