// internal/task_lifecycle.go
package internal

import "fmt"

// validChangeTransitions defines the allowed state transitions for Changes.
var validChangeTransitions = map[ChangeStatus][]ChangeStatus{
	ChangeStatusDraft:        {ChangeStatusAssigned, ChangeStatusArchived},
	ChangeStatusAssigned:     {ChangeStatusImplementing, ChangeStatusDraft, ChangeStatusArchived},
	ChangeStatusImplementing: {ChangeStatusVerifying, ChangeStatusAssigned, ChangeStatusArchived},
	ChangeStatusVerifying:    {ChangeStatusDone, ChangeStatusImplementing, ChangeStatusArchived},
	ChangeStatusDone:         {ChangeStatusArchived},
	ChangeStatusArchived:     {}, // terminal state
}

// ValidateChangeTransition returns an error if the transition from→to is not allowed.
func ValidateChangeTransition(from, to ChangeStatus) error {
	allowed, ok := validChangeTransitions[from]
	if !ok {
		return fmt.Errorf("unknown status: %s", from)
	}

	for _, a := range allowed {
		if a == to {
			return nil // transition is valid
		}
	}

	return fmt.Errorf("invalid transition: %s → %s", from, to)
}

// ApplyChangeTransition updates the change's status if the transition is valid.
func ApplyChangeTransition(change *Change, to ChangeStatus) error {
	if err := ValidateChangeTransition(change.Status, to); err != nil {
		return err
	}
	change.Status = to
	return nil
}

// IsTerminalChangeStatus returns true if the status is terminal (no further transitions allowed).
func IsTerminalChangeStatus(s ChangeStatus) bool {
	transitions, ok := validChangeTransitions[s]
	return ok && len(transitions) == 0
}

// AllTasksDone returns true if all tasks in the change are done or skipped.
// Empty task list returns true.
func AllTasksDone(change *Change) bool {
	if len(change.Tasks) == 0 {
		return true
	}

	for _, task := range change.Tasks {
		if task.Status != TaskStatusDone && task.Status != TaskStatusSkipped {
			return false
		}
	}

	return true
}

// AutoTransitionOnVerify automatically transitions a change based on verification result.
// If passed=true and change is in verifying state, transitions to done.
// If passed=false and change is in verifying state, transitions back to implementing.
func AutoTransitionOnVerify(change *Change, passed bool) error {
	if change.Status != ChangeStatusVerifying {
		return fmt.Errorf("can only auto-transition from verifying state, current: %s", change.Status)
	}

	if passed {
		return ApplyChangeTransition(change, ChangeStatusDone)
	}

	return ApplyChangeTransition(change, ChangeStatusImplementing)
}
