package governance

import (
	"testing"
	"time"
)

func TestArchivedExceptionConsumeOneTimeReadOnly(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	ticket := NewArchivedExceptionTicket("tk-1", "task-1", "owner-1", "debug", now)

	if !ticket.ReadOnly || !ticket.SingleTask {
		t.Fatalf("ticket should be read-only and single-task")
	}

	if err := ConsumeArchivedException(&ticket, "task-1", "owner-1", now); err != nil {
		t.Fatalf("consume failed: %v", err)
	}
	if ticket.Status != ArchivedExceptionStatusUsed || ticket.UsedAt == nil {
		t.Fatalf("ticket should be used after first consume")
	}

	if err := ConsumeArchivedException(&ticket, "task-1", "owner-1", now); err == nil {
		t.Fatalf("second consume should fail")
	}
}
