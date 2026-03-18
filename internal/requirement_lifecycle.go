// internal/requirement_lifecycle.go
package internal

import "fmt"

// validSubTaskTransitions defines allowed state transitions for SubTasks.
var validSubTaskTransitions = map[SubTaskStatus][]SubTaskStatus{
	SubTaskStatusPending:  {SubTaskStatusAssigned, SubTaskStatusSkipped},
	SubTaskStatusAssigned: {SubTaskStatusDone, SubTaskStatusSkipped, SubTaskStatusPending},
	SubTaskStatusDone:     {},
	SubTaskStatusSkipped:  {},
}

// ValidateSubTaskTransition returns an error if the transition from→to is not allowed.
func ValidateSubTaskTransition(from, to SubTaskStatus) error {
	allowed, ok := validSubTaskTransitions[from]
	if !ok {
		return fmt.Errorf("unknown sub-task status: %s", from)
	}

	for _, a := range allowed {
		if a == to {
			return nil
		}
	}

	return fmt.Errorf("invalid sub-task transition: %s → %s", from, to)
}

// AutoPromoteRequirement checks if all sub-tasks are done/skipped and
// automatically sets the requirement status to done.
// Returns true if the requirement was promoted.
func AutoPromoteRequirement(req *Requirement) bool {
	if req.Status == RequirementStatusDone {
		return false
	}

	if len(req.SubTasks) == 0 {
		return false
	}

	for _, st := range req.SubTasks {
		if st.Status != SubTaskStatusDone && st.Status != SubTaskStatusSkipped {
			return false
		}
	}

	req.Status = RequirementStatusDone
	return true
}

// MarkSubTaskDone marks a sub-task as done and auto-promotes the requirement if all are done.
func MarkSubTaskDone(req *Requirement, subTaskID int) error {
	idx := -1
	for i := range req.SubTasks {
		if req.SubTasks[i].ID == subTaskID {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("sub-task %d not found", subTaskID)
	}

	if err := ValidateSubTaskTransition(req.SubTasks[idx].Status, SubTaskStatusDone); err != nil {
		return err
	}

	req.SubTasks[idx].Status = SubTaskStatusDone
	AutoPromoteRequirement(req)

	return nil
}

// AssignSubTask assigns a sub-task to a worker, setting assigned_to and change_name.
// Transitions SubTask from pending→assigned and Requirement from open→in_progress.
func AssignSubTask(req *Requirement, subTaskID int, workerID, changeName string) error {
	idx := -1
	for i := range req.SubTasks {
		if req.SubTasks[i].ID == subTaskID {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("sub-task %d not found", subTaskID)
	}

	if err := ValidateSubTaskTransition(req.SubTasks[idx].Status, SubTaskStatusAssigned); err != nil {
		return err
	}

	req.SubTasks[idx].Status = SubTaskStatusAssigned
	req.SubTasks[idx].AssignedTo = workerID
	req.SubTasks[idx].ChangeName = changeName

	if req.Status == RequirementStatusOpen {
		req.Status = RequirementStatusInProgress
	}

	return nil
}
