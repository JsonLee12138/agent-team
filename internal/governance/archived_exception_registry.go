package governance

import "fmt"

type InMemoryArchivedExceptionRegistry struct {
	tickets map[string]ArchivedExceptionTicket
}

func NewInMemoryArchivedExceptionRegistry() *InMemoryArchivedExceptionRegistry {
	return &InMemoryArchivedExceptionRegistry{tickets: make(map[string]ArchivedExceptionTicket)}
}

func (r *InMemoryArchivedExceptionRegistry) Issue(ticket ArchivedExceptionTicket) error {
	if ticket.TicketID == "" {
		return fmt.Errorf("ticket id is required")
	}
	if _, exists := r.tickets[ticket.TicketID]; exists {
		return fmt.Errorf("ticket already exists: %s", ticket.TicketID)
	}
	r.tickets[ticket.TicketID] = ticket
	return nil
}

func (r *InMemoryArchivedExceptionRegistry) Consume(ticketID, taskID, owner string) error {
	ticket, ok := r.tickets[ticketID]
	if !ok {
		return fmt.Errorf("ticket not found: %s", ticketID)
	}
	if err := ConsumeArchivedException(&ticket, taskID, owner, TimeNowUTC()); err != nil {
		return err
	}
	r.tickets[ticketID] = ticket
	return nil
}

func (r *InMemoryArchivedExceptionRegistry) Get(ticketID string) (*ArchivedExceptionTicket, bool) {
	ticket, ok := r.tickets[ticketID]
	if !ok {
		return nil, false
	}
	copied := ticket
	return &copied, true
}
