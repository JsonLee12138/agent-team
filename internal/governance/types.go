package governance

import "time"

const (
	GateLevelBlocker = "blocker"
	GateLevelInfo    = "info"
)

const (
	WorkflowPlanStatusProposed = "proposed"
	WorkflowPlanStatusApproved = "approved"
	WorkflowPlanStatusActive   = "active"
	WorkflowPlanStatusClosed   = "closed"
)

// TaskPacket is the canonical governance input envelope for task-level decisions.
type TaskPacket struct {
	TaskID             string   `yaml:"task_id"`
	ModuleID           string   `yaml:"module_id,omitempty"`
	Owner              string   `yaml:"owner"`
	Actor              string   `yaml:"actor,omitempty"`
	DeclaredReferences []string `yaml:"declared_references,omitempty"`
	UsesArchivedInput  bool     `yaml:"uses_archived_input,omitempty"`
}

// IndexEntry is a normalized pointer entry used by Index-First checks.
type IndexEntry struct {
	ID       string `yaml:"id"`
	Kind     string `yaml:"kind,omitempty"`
	Path     string `yaml:"path,omitempty"`
	Archived bool   `yaml:"archived,omitempty"`
}

// Index keeps summary pointers. Bodies live in module-level storage.
type Index struct {
	Entries []IndexEntry `yaml:"entries,omitempty"`
}

// GateResult is the fixed gate output shape.
type GateResult struct {
	Code       string            `yaml:"code"`
	Level      string            `yaml:"level"`
	Message    string            `yaml:"message"`
	Context    map[string]string `yaml:"context,omitempty"`
	NextAction string            `yaml:"next_action,omitempty"`
}

func (g GateResult) IsBlocker() bool {
	return g.Level == GateLevelBlocker
}

func PassGateResult() GateResult {
	return GateResult{
		Code:       "ok",
		Level:      GateLevelInfo,
		Message:    "gate checks passed",
		Context:    map[string]string{},
		NextAction: "continue",
	}
}

// WorkflowPlan is the governance-owned orchestration object.
type WorkflowPlan struct {
	ID        string    `yaml:"id"`
	TaskID    string    `yaml:"task_id"`
	Owner     string    `yaml:"owner"`
	Status    string    `yaml:"status"`
	InputRefs []string  `yaml:"input_refs,omitempty"`
	Reasons   []string  `yaml:"reasons,omitempty"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
}

// Rule models a single text rule item.
type Rule struct {
	ID    string `yaml:"id"`
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

// RuleConflict captures lower-priority override attempts.
type RuleConflict struct {
	Key            string
	HigherRuleID   string
	HigherRuleFrom string
	LowerRuleID    string
	LowerRuleFrom  string
}

// RuleLoadResult contains effective rules and conflicts.
type RuleLoadResult struct {
	Effective []Rule
	Conflicts []RuleConflict
}

// AdvisorInput contains only text-artifact-based inputs.
type AdvisorInput struct {
	PlanID       string
	TaskPacket   TaskPacket
	Index        Index
	EvidenceRefs []string
	Reasons      []string
	Now          time.Time
}
