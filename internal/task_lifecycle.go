package internal

import "fmt"

var validTaskTransitions = map[TaskStatus][]TaskStatus{
	TaskStatusDraft:    {TaskStatusAssigned},
	TaskStatusAssigned: {TaskStatusDone},
	TaskStatusDone:     {TaskStatusAssigned, TaskStatusArchived},
	TaskStatusArchived: {},
}

func ValidateTaskTransition(from, to TaskStatus) error {
	allowed, ok := validTaskTransitions[from]
	if !ok {
		return fmt.Errorf("unknown task status: %s", from)
	}
	for _, next := range allowed {
		if next == to {
			return nil
		}
	}
	return fmt.Errorf("invalid task transition: %s → %s", from, to)
}
