package governance

// IndexProvider abstracts Index-First data access.
type IndexProvider interface {
	LoadIndex() (Index, error)
}

// WorkflowPlanStore abstracts WorkflowPlan persistence.
type WorkflowPlanStore interface {
	SaveWorkflowPlan(plan *WorkflowPlan) error
	LoadWorkflowPlan(planID string) (*WorkflowPlan, error)
}

// ArchivedExceptionRegistry abstracts exception ticket lifecycle.
type ArchivedExceptionRegistry interface {
	Issue(ticket ArchivedExceptionTicket) error
	Consume(ticketID, taskID, owner string) error
}
