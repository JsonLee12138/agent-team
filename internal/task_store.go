package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var saveTaskRecordAt = saveTaskRecordAtImpl

func AgentTeamDir(root string) string {
	return filepath.Join(root, ".agent-team")
}

func TasksRootDir(root string) string {
	return filepath.Join(AgentTeamDir(root), "task")
}

func TasksArchiveRootDir(root string) string {
	return filepath.Join(AgentTeamDir(root), "archive", "task")
}

func TasksDeprecatedRootDir(root string) string {
	return filepath.Join(AgentTeamDir(root), "deprecated", "task")
}

func TaskDir(root, taskID string) string {
	return filepath.Join(TasksRootDir(root), taskID)
}

func TaskArchiveDir(root, taskID string) string {
	return filepath.Join(TasksArchiveRootDir(root), taskID)
}

func TaskDeprecatedDir(root, taskID string) string {
	return filepath.Join(TasksDeprecatedRootDir(root), taskID)
}

func TaskYAMLPath(root, taskID string) string {
	return filepath.Join(TaskDir(root, taskID), "task.yaml")
}

func TaskContextPath(root, taskID string) string {
	return filepath.Join(TaskDir(root, taskID), "context.md")
}

func TaskVerificationPath(root, taskID string) string {
	return filepath.Join(TaskDir(root, taskID), "verification.md")
}

func TaskArchiveYAMLPath(root, taskID string) string {
	return filepath.Join(TaskArchiveDir(root, taskID), "task.yaml")
}

func TaskArchiveContextPath(root, taskID string) string {
	return filepath.Join(TaskArchiveDir(root, taskID), "context.md")
}

func TaskArchiveVerificationPath(root, taskID string) string {
	return filepath.Join(TaskArchiveDir(root, taskID), "verification.md")
}

func TaskDeprecatedYAMLPath(root, taskID string) string {
	return filepath.Join(TaskDeprecatedDir(root, taskID), "task.yaml")
}

func TaskDeprecatedContextPath(root, taskID string) string {
	return filepath.Join(TaskDeprecatedDir(root, taskID), "context.md")
}

func TaskDeprecatedVerificationPath(root, taskID string) string {
	return filepath.Join(TaskDeprecatedDir(root, taskID), "verification.md")
}

func TaskRelPath(taskID string) string {
	return filepath.ToSlash(filepath.Join(".agent-team", "task", taskID))
}

func TaskArchiveRelPath(taskID string) string {
	return filepath.ToSlash(filepath.Join(".agent-team", "archive", "task", taskID))
}

func TaskDeprecatedRelPath(taskID string) string {
	return filepath.ToSlash(filepath.Join(".agent-team", "deprecated", "task", taskID))
}

func GenerateTaskID(title string, now time.Time) string {
	return fmt.Sprintf("%s-%s", now.UTC().Format("2006-01-02-15-04-05"), Slugify(title, 48))
}

func CreateTaskPackage(root, title, role, design string, now time.Time) (*TaskRecord, error) {
	taskID := GenerateTaskID(title, now)
	record := &TaskRecord{
		TaskID:    taskID,
		Title:     title,
		Role:      role,
		Status:    TaskStatusDraft,
		TaskPath:  TaskRelPath(taskID),
		CreatedAt: now.UTC().Format(time.RFC3339),
	}

	taskDir := TaskDir(root, taskID)
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return nil, fmt.Errorf("create task directory: %w", err)
	}
	if err := saveTaskRecordAt(taskDir, record); err != nil {
		return nil, err
	}
	if err := os.WriteFile(TaskContextPath(root, taskID), []byte(defaultTaskContext(record, design)), 0644); err != nil {
		return nil, fmt.Errorf("write context.md: %w", err)
	}
	if err := os.WriteFile(TaskVerificationPath(root, taskID), []byte(defaultTaskVerification()), 0644); err != nil {
		return nil, fmt.Errorf("write verification.md: %w", err)
	}
	return record, nil
}

func LoadTaskRecord(root, taskID string) (*TaskRecord, TaskRecordLocation, error) {
	if record, err := loadTaskRecordFromDir(TaskDir(root, taskID)); err == nil {
		return record, TaskRecordLocationActive, nil
	}
	if record, err := loadTaskRecordFromDir(TaskArchiveDir(root, taskID)); err == nil {
		return record, TaskRecordLocationArchived, nil
	}
	record, err := loadTaskRecordFromDir(TaskDeprecatedDir(root, taskID))
	if err != nil {
		return nil, "", err
	}
	return record, TaskRecordLocationDeprecated, nil
}

func SaveTaskRecord(root string, record *TaskRecord) error {
	if record == nil {
		return fmt.Errorf("task record is nil")
	}
	record.LegacyDoneAt = ""
	var dir string
	switch record.Status {
	case TaskStatusArchived:
		dir = TaskArchiveDir(root, record.TaskID)
	case TaskStatusDeprecated:
		dir = TaskDeprecatedDir(root, record.TaskID)
	default:
		dir = TaskDir(root, record.TaskID)
	}
	return saveTaskRecordAt(dir, record)
}

func ListTasks(root string, activeOnly bool) ([]*TaskRecord, error) {
	var dirs []string
	dirs = append(dirs, TasksRootDir(root))
	if !activeOnly {
		dirs = append(dirs, TasksArchiveRootDir(root), TasksDeprecatedRootDir(root))
	}

	var tasks []*TaskRecord
	for _, base := range dirs {
		entries, err := os.ReadDir(base)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read task directory %s: %w", base, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			record, err := loadTaskRecordFromDir(filepath.Join(base, entry.Name()))
			if err != nil {
				continue
			}
			if activeOnly && (record.Status == TaskStatusArchived || record.Status == TaskStatusDeprecated) {
				continue
			}
			tasks = append(tasks, record)
		}
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].TaskID < tasks[j].TaskID
	})
	return tasks, nil
}

func BindTaskToWorker(root, taskID, workerID string, now time.Time) (*TaskRecord, error) {
	record, location, err := LoadTaskRecord(root, taskID)
	if err != nil {
		return nil, err
	}
	if location != TaskRecordLocationActive {
		return nil, fmt.Errorf("task '%s' is %s", taskID, location)
	}

	switch record.Status {
	case TaskStatusDraft, TaskStatusVerifying:
		if err := ValidateTaskTransition(record.Status, TaskStatusAssigned); err != nil {
			return nil, err
		}
	case TaskStatusAssigned:
		// Keep assigned status for reopen/rebind flows.
	default:
		return nil, fmt.Errorf("task '%s' cannot be assigned from status '%s'", taskID, record.Status)
	}

	record.Status = TaskStatusAssigned
	record.WorkerID = workerID
	record.TaskPath = TaskRelPath(taskID)
	record.AssignedAt = now.UTC().Format(time.RFC3339)
	record.VerifyingAt = ""
	record.ArchivedAt = ""
	record.DeprecatedAt = ""
	record.MergedSHA = ""
	if err := SaveTaskRecord(root, record); err != nil {
		return nil, err
	}
	return record, nil
}

func MarkTaskDone(root, taskID string, now time.Time) (*TaskRecord, error) {
	record, location, err := LoadTaskRecord(root, taskID)
	if err != nil {
		return nil, err
	}
	if location != TaskRecordLocationActive {
		return nil, fmt.Errorf("task '%s' is %s", taskID, location)
	}
	if err := ValidateTaskTransition(record.Status, TaskStatusVerifying); err != nil {
		return nil, err
	}
	record.Status = TaskStatusVerifying
	record.VerifyingAt = now.UTC().Format(time.RFC3339)
	if err := SaveTaskRecord(root, record); err != nil {
		return nil, err
	}
	return record, nil
}

func ArchiveTask(root, taskID, mergedSHA string, strict bool, now time.Time) (*TaskRecord, error) {
	if strings.TrimSpace(mergedSHA) == "" {
		return nil, fmt.Errorf("merged SHA is required")
	}
	record, location, err := LoadTaskRecord(root, taskID)
	if err != nil {
		return nil, err
	}
	if location == TaskRecordLocationArchived {
		return nil, fmt.Errorf("task '%s' is already archived", taskID)
	}
	if location == TaskRecordLocationDeprecated {
		return nil, fmt.Errorf("task '%s' is deprecated", taskID)
	}
	if err := ValidateTaskTransition(record.Status, TaskStatusArchived); err != nil {
		return nil, err
	}
	result, err := ReadTaskVerificationResult(root, taskID, TaskRecordLocationActive)
	if err != nil {
		return nil, err
	}
	if err := ValidateArchiveReadiness(result, strict); err != nil {
		return nil, err
	}

	record.Status = TaskStatusArchived
	record.TaskPath = TaskArchiveRelPath(taskID)
	record.ArchivedAt = now.UTC().Format(time.RFC3339)
	record.DeprecatedAt = ""
	record.MergedSHA = mergedSHA
	return moveTaskPackage(root, record, TaskRecordLocationActive, TaskRecordLocationArchived)
}

func DeprecateTask(root, taskID string, now time.Time) (*TaskRecord, error) {
	record, location, err := LoadTaskRecord(root, taskID)
	if err != nil {
		return nil, err
	}
	if location == TaskRecordLocationArchived {
		return nil, fmt.Errorf("task '%s' is already archived", taskID)
	}
	if location == TaskRecordLocationDeprecated {
		return nil, fmt.Errorf("task '%s' is already deprecated", taskID)
	}
	if err := ValidateTaskTransition(record.Status, TaskStatusDeprecated); err != nil {
		return nil, err
	}

	record.Status = TaskStatusDeprecated
	record.TaskPath = TaskDeprecatedRelPath(taskID)
	record.DeprecatedAt = now.UTC().Format(time.RFC3339)
	record.ArchivedAt = ""
	record.MergedSHA = ""
	return moveTaskPackage(root, record, TaskRecordLocationActive, TaskRecordLocationDeprecated)
}

func moveTaskPackage(root string, record *TaskRecord, from, to TaskRecordLocation) (*TaskRecord, error) {
	srcDir := taskDirByLocation(root, record.TaskID, from)
	dstDir := taskDirByLocation(root, record.TaskID, to)
	if err := os.MkdirAll(filepath.Dir(dstDir), 0755); err != nil {
		return nil, fmt.Errorf("create destination parent directory: %w", err)
	}
	if _, err := os.Stat(dstDir); err == nil {
		return nil, fmt.Errorf("destination already exists: %s", dstDir)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("check destination: %w", err)
	}
	if err := os.Rename(srcDir, dstDir); err != nil {
		return nil, fmt.Errorf("move task package: %w", err)
	}
	if err := saveTaskRecordAt(dstDir, record); err != nil {
		if rollbackErr := os.Rename(dstDir, srcDir); rollbackErr != nil {
			return nil, fmt.Errorf("write moved task metadata: %w (rollback failed: %v)", err, rollbackErr)
		}
		return nil, fmt.Errorf("write moved task metadata: %w", err)
	}
	return record, nil
}

func taskDirByLocation(root, taskID string, location TaskRecordLocation) string {
	switch location {
	case TaskRecordLocationArchived:
		return TaskArchiveDir(root, taskID)
	case TaskRecordLocationDeprecated:
		return TaskDeprecatedDir(root, taskID)
	default:
		return TaskDir(root, taskID)
	}
}

func loadTaskRecordFromDir(dir string) (*TaskRecord, error) {
	data, err := os.ReadFile(filepath.Join(dir, "task.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read task.yaml: %w", err)
	}
	var record TaskRecord
	if err := yaml.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("parse task.yaml: %w", err)
	}
	if record.Status == TaskStatus("done") {
		record.Status = TaskStatusVerifying
		if record.VerifyingAt == "" {
			record.VerifyingAt = record.LegacyDoneAt
		}
	}
	if !ValidTaskStatus(record.Status) {
		return nil, fmt.Errorf("invalid task status: %s", record.Status)
	}
	return &record, nil
}

func saveTaskRecordAtImpl(dir string, record *TaskRecord) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create task directory: %w", err)
	}
	record.LegacyDoneAt = ""
	data, err := yaml.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal task record: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "task.yaml"), data, 0644); err != nil {
		return fmt.Errorf("write task.yaml: %w", err)
	}
	return nil
}

func defaultTaskContext(record *TaskRecord, design string) string {
	if strings.TrimSpace(design) == "" {
		design = "- None provided yet."
	}
	return fmt.Sprintf("# Task Context\n\n- Task ID: `%s`\n- Title: %s\n- Role: `%s`\n\n## Background\n\n- TODO\n\n## Scope\n\n- TODO\n\n## Acceptance\n\n- TODO\n\n## Constraints\n\n- Follow the approved workflow and repository rules.\n- Keep the task scoped to this assignment.\n\n## Design\n\n%s\n", record.TaskID, record.Title, record.Role, strings.TrimSpace(design))
}

func defaultTaskVerification() string {
	return "# Verification\n\n## Acceptance Criteria\n- TODO\n\n## Test Scope\n- Unit Test Coverage Required: yes\n- E2E Required: no\n\n## Checks Performed\n- Not run yet.\n\n## Result\n- pending\n\n## Issues\n- None.\n\n## Verified By\n- qa\n\n## Verified At\n- TODO\n"
}

// Legacy .tasks helpers are kept for read-only compatibility during migration.
func TasksDir(wtPath string) string {
	return filepath.Join(wtPath, ".tasks")
}

func TasksChangesDir(wtPath string) string {
	return filepath.Join(TasksDir(wtPath), "changes")
}

func ChangeDirPath(wtPath, changeName string) string {
	return filepath.Join(TasksChangesDir(wtPath), changeName)
}

func CreateTaskChange(wtPath, changeName, description, proposal, design string) (string, error) {
	_ = description
	changeDir := ChangeDirPath(wtPath, changeName)
	if err := os.MkdirAll(changeDir, 0755); err != nil {
		return "", fmt.Errorf("create change directory: %w", err)
	}
	if proposal != "" {
		proposalPath := filepath.Join(changeDir, "proposal.md")
		if err := os.WriteFile(proposalPath, []byte(proposal), 0644); err != nil {
			return "", fmt.Errorf("write proposal.md: %w", err)
		}
	}
	if design != "" {
		designPath := filepath.Join(changeDir, "design.md")
		if err := os.WriteFile(designPath, []byte(design), 0644); err != nil {
			return "", fmt.Errorf("write design.md: %w", err)
		}
	}
	return changeDir, nil
}

func ListChanges(wtPath string) ([]string, error) {
	changesDir := TasksChangesDir(wtPath)
	if _, err := os.Stat(changesDir); os.IsNotExist(err) {
		return []string{}, nil
	}
	entries, err := os.ReadDir(changesDir)
	if err != nil {
		return nil, fmt.Errorf("read changes directory: %w", err)
	}
	var changes []string
	for _, e := range entries {
		if e.IsDir() {
			changes = append(changes, e.Name())
		}
	}
	sort.Strings(changes)
	return changes, nil
}

func CountActiveChanges(wtPath string) int {
	changes, err := ListChanges(wtPath)
	if err != nil {
		return 0
	}
	return len(changes)
}

func InitTasksDir(wtPath string) error {
	changesDir := TasksChangesDir(wtPath)
	if err := os.MkdirAll(changesDir, 0755); err != nil {
		return fmt.Errorf("create tasks directories: %w", err)
	}
	return nil
}
