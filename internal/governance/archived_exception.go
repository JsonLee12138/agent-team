package governance

import (
	"fmt"
	"time"
)

const (
	ArchivedExceptionStatusIssued = "issued"
	ArchivedExceptionStatusUsed   = "used"
)

type ArchivedExceptionTicket struct {
	TicketID   string     `yaml:"ticket_id"`
	TaskID     string     `yaml:"task_id"`
	Owner      string     `yaml:"owner"`
	Reason     string     `yaml:"reason"`
	CreatedAt  time.Time  `yaml:"created_at"`
	UsedAt     *time.Time `yaml:"used_at,omitempty"`
	Status     string     `yaml:"status"`
	ReadOnly   bool       `yaml:"read_only"`
	SingleTask bool       `yaml:"single_task"`
}

func NewArchivedExceptionTicket(ticketID, taskID, owner, reason string, now time.Time) ArchivedExceptionTicket {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return ArchivedExceptionTicket{
		TicketID:   ticketID,
		TaskID:     taskID,
		Owner:      owner,
		Reason:     reason,
		CreatedAt:  now,
		Status:     ArchivedExceptionStatusIssued,
		ReadOnly:   true,
		SingleTask: true,
	}
}

func ConsumeArchivedException(ticket *ArchivedExceptionTicket, taskID, owner string, now time.Time) error {
	if ticket == nil {
		return fmt.Errorf("archived exception ticket is required")
	}
	if !ticket.ReadOnly {
		return fmt.Errorf("archived exception must be read-only")
	}
	if !ticket.SingleTask {
		return fmt.Errorf("archived exception must be single-task scoped")
	}
	if ticket.TaskID != taskID {
		return fmt.Errorf("ticket task mismatch: want %s, got %s", taskID, ticket.TaskID)
	}
	if ticket.Owner != owner {
		return fmt.Errorf("ticket owner mismatch: want %s, got %s", owner, ticket.Owner)
	}
	if ticket.Status != ArchivedExceptionStatusIssued {
		return fmt.Errorf("ticket is not available: status=%s", ticket.Status)
	}
	if ticket.UsedAt != nil {
		return fmt.Errorf("ticket already used")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	ticket.UsedAt = &now
	ticket.Status = ArchivedExceptionStatusUsed
	return nil
}
