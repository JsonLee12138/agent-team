package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	WorkflowStatusReady               = "ready"
	WorkflowStatusRunning             = "running"
	WorkflowStatusWaiting             = "waiting"
	WorkflowStatusWaitingConfirmation = "waiting_confirmation"
	WorkflowStatusBlocked             = "blocked"
	WorkflowStatusCompleted           = "completed"
)

const (
	WorkflowNodeStatusPending   = "pending"
	WorkflowNodeStatusRunning   = "running"
	WorkflowNodeStatusWaiting   = "waiting"
	WorkflowNodeStatusCompleted = "completed"
	WorkflowNodeStatusBlocked   = "blocked"
)

var supportedWorkflowNodeTypes = map[string]struct{}{
	"controller_task":     {},
	"assign_role_task":    {},
	"wait_for_completion": {},
	"decision":            {},
	"handoff":             {},
	"verify_or_test":      {},
	"merge":               {},
}

var actorRequiredWorkflowNodeTypes = map[string]struct{}{
	"assign_role_task":    {},
	"wait_for_completion": {},
	"handoff":             {},
	"verify_or_test":      {},
	"merge":               {},
}

type WorkflowDefaults struct {
	ExecutionMode         string `yaml:"execution_mode,omitempty"`
	CreateWorkerIfMissing bool   `yaml:"create_worker_if_missing,omitempty"`
	ReuseWorker           bool   `yaml:"reuse_worker,omitempty"`
}

type WorkflowRole struct {
	Role   string `yaml:"role"`
	Worker string `yaml:"worker,omitempty"`
}

type WorkflowBranch struct {
	When string `yaml:"when"`
	Next string `yaml:"next"`
}

type WorkflowNode struct {
	ID                   string           `yaml:"id"`
	Type                 string           `yaml:"type"`
	Actor                string           `yaml:"actor,omitempty"`
	Task                 string           `yaml:"task,omitempty"`
	Next                 string           `yaml:"next,omitempty"`
	Branches             []WorkflowBranch `yaml:"branches,omitempty"`
	RequiresConfirmation bool             `yaml:"requires_confirmation,omitempty"`
	End                  bool             `yaml:"end,omitempty"`
}

type WorkflowTemplate struct {
	Version  int                     `yaml:"version"`
	Name     string                  `yaml:"name"`
	Roles    map[string]WorkflowRole `yaml:"roles"`
	Defaults WorkflowDefaults        `yaml:"defaults,omitempty"`
	Entry    string                  `yaml:"entry"`
	Nodes    []WorkflowNode          `yaml:"nodes"`
}

type WorkflowNodeState struct {
	Status      string `yaml:"status"`
	StartedAt   string `yaml:"started_at,omitempty"`
	CompletedAt string `yaml:"completed_at,omitempty"`
	Outcome     string `yaml:"outcome,omitempty"`
	Worker      string `yaml:"worker,omitempty"`
	Summary     string `yaml:"summary,omitempty"`
}

type WorkflowDecision struct {
	Timestamp string `yaml:"timestamp"`
	Node      string `yaml:"node"`
	Action    string `yaml:"action"`
	Outcome   string `yaml:"outcome,omitempty"`
	NextNode  string `yaml:"next_node,omitempty"`
	Summary   string `yaml:"summary,omitempty"`
}

type WorkflowRunState struct {
	Version             int                          `yaml:"version"`
	RunID               string                       `yaml:"run_id"`
	WorkflowFile        string                       `yaml:"workflow_file"`
	WorkflowName        string                       `yaml:"workflow_name"`
	Status              string                       `yaml:"status"`
	CurrentNode         string                       `yaml:"current_node,omitempty"`
	PendingConfirmation string                       `yaml:"pending_confirmation,omitempty"`
	BlockingReason      string                       `yaml:"blocking_reason,omitempty"`
	RoleWorkerMap       map[string]string            `yaml:"role_worker_map,omitempty"`
	NodeStates          map[string]WorkflowNodeState `yaml:"node_states"`
	DecisionLog         []WorkflowDecision           `yaml:"decision_log,omitempty"`
	CreatedAt           string                       `yaml:"created_at"`
	UpdatedAt           string                       `yaml:"updated_at"`
}

func WorkflowDir(root string) string {
	return filepath.Join(root, ".agents", "workflow")
}

func WorkflowTemplatePath(root, name string) string {
	return filepath.Join(WorkflowDir(root), fmt.Sprintf("%s.yaml", name))
}

func WorkflowRunPath(root, workflowName, runID string) string {
	return filepath.Join(WorkflowDir(root), "runs", workflowName, fmt.Sprintf("%s.yaml", runID))
}

func LoadWorkflowTemplate(path string) (*WorkflowTemplate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow template %s: %w", path, err)
	}
	var wf WorkflowTemplate
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("parse workflow template %s: %w", path, err)
	}
	return &wf, nil
}

func SaveWorkflowTemplate(path string, wf *WorkflowTemplate) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create workflow template directory: %w", err)
	}
	data, err := yaml.Marshal(wf)
	if err != nil {
		return fmt.Errorf("marshal workflow template: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write workflow template %s: %w", path, err)
	}
	return nil
}

func ValidateWorkflowTemplate(wf *WorkflowTemplate) []string {
	var errs []string
	if wf.Version == 0 {
		errs = append(errs, "missing top-level field: version")
	}
	if strings.TrimSpace(wf.Name) == "" {
		errs = append(errs, "missing top-level field: name")
	}
	if len(wf.Roles) == 0 {
		errs = append(errs, "missing top-level field: roles")
	}
	if strings.TrimSpace(wf.Entry) == "" {
		errs = append(errs, "missing top-level field: entry")
	}
	if len(wf.Nodes) == 0 {
		errs = append(errs, "missing top-level field: nodes")
	}
	if len(errs) > 0 {
		return errs
	}

	nodeMap := make(map[string]WorkflowNode, len(wf.Nodes))
	for idx, node := range wf.Nodes {
		if strings.TrimSpace(node.ID) == "" {
			errs = append(errs, fmt.Sprintf("nodes[%d] is missing a valid id", idx))
			continue
		}
		if _, exists := nodeMap[node.ID]; exists {
			errs = append(errs, fmt.Sprintf("duplicate node id: %s", node.ID))
			continue
		}
		nodeMap[node.ID] = node

		if _, ok := supportedWorkflowNodeTypes[node.Type]; !ok {
			errs = append(errs, fmt.Sprintf("node '%s' has unsupported type: %s", node.ID, node.Type))
		}
		if _, actorRequired := actorRequiredWorkflowNodeTypes[node.Type]; actorRequired {
			if _, exists := wf.Roles[node.Actor]; !exists || strings.TrimSpace(node.Actor) == "" {
				errs = append(errs, fmt.Sprintf("node '%s' requires actor mapped in roles, got: %s", node.ID, node.Actor))
			}
		}
		if !node.End && strings.TrimSpace(node.Next) == "" && len(node.Branches) == 0 {
			errs = append(errs, fmt.Sprintf("node '%s' must define next, branches, or end: true", node.ID))
		}
		for branchIdx, branch := range node.Branches {
			if strings.TrimSpace(branch.When) == "" {
				errs = append(errs, fmt.Sprintf("node '%s' branches[%d] missing valid when", node.ID, branchIdx))
			}
			if strings.TrimSpace(branch.Next) == "" {
				errs = append(errs, fmt.Sprintf("node '%s' branches[%d] missing valid next", node.ID, branchIdx))
			}
		}
	}

	if _, exists := nodeMap[wf.Entry]; !exists {
		errs = append(errs, fmt.Sprintf("entry node not found: %s", wf.Entry))
	}

	for _, node := range wf.Nodes {
		if node.Next != "" {
			if _, exists := nodeMap[node.Next]; !exists {
				errs = append(errs, fmt.Sprintf("node '%s' points to undefined next target: %s", node.ID, node.Next))
			}
		}
		for _, branch := range node.Branches {
			if branch.Next == "" {
				continue
			}
			if _, exists := nodeMap[branch.Next]; !exists {
				errs = append(errs, fmt.Sprintf("node '%s' branch '%s' points to undefined target: %s", node.ID, branch.When, branch.Next))
			}
		}
	}

	return errs
}

func NewWorkflowTemplate(name, preset, ctoRole, devRole, qaRole, executionMode string) (*WorkflowTemplate, error) {
	var nodes []WorkflowNode
	switch preset {
	case "dev-first":
		nodes = []WorkflowNode{
			{ID: "cto_breakdown", Type: "assign_role_task", Actor: "cto", Task: "拆分需求并给出实施子任务", Next: "dev_implement"},
			{ID: "dev_implement", Type: "assign_role_task", Actor: "dev", Task: "根据拆分结果实现功能", Next: "qa_verify"},
			{ID: "qa_verify", Type: "verify_or_test", Actor: "qa", Task: "执行验证并反馈结果", Branches: []WorkflowBranch{{When: "passed", Next: "controller_finish"}, {When: "failed", Next: "dev_fix"}}},
			{ID: "dev_fix", Type: "assign_role_task", Actor: "dev", Task: "根据测试反馈修复问题", Next: "qa_verify"},
			{ID: "controller_finish", Type: "controller_task", Task: "决定是否合并与关闭流程", RequiresConfirmation: true, End: true},
		}
	case "test-first":
		nodes = []WorkflowNode{
			{ID: "cto_breakdown", Type: "assign_role_task", Actor: "cto", Task: "拆分需求并给出实施子任务", Next: "qa_write_tests"},
			{ID: "qa_write_tests", Type: "assign_role_task", Actor: "qa", Task: "先编写验收测试或测试用例", Next: "dev_implement"},
			{ID: "dev_implement", Type: "assign_role_task", Actor: "dev", Task: "根据需求与测试实现功能", Next: "qa_verify"},
			{ID: "qa_verify", Type: "verify_or_test", Actor: "qa", Task: "执行验证并反馈结果", Branches: []WorkflowBranch{{When: "passed", Next: "controller_finish"}, {When: "failed", Next: "dev_fix"}}},
			{ID: "dev_fix", Type: "assign_role_task", Actor: "dev", Task: "根据测试反馈修复问题", Next: "qa_verify"},
			{ID: "controller_finish", Type: "controller_task", Task: "决定是否合并与关闭流程", RequiresConfirmation: true, End: true},
		}
	case "branching":
		nodes = []WorkflowNode{
			{ID: "cto_breakdown", Type: "assign_role_task", Actor: "cto", Task: "拆分需求并给出实施子任务", Next: "controller_dispatch"},
			{ID: "controller_dispatch", Type: "controller_task", Task: "审阅拆分结果并决定分发路径", RequiresConfirmation: true, Branches: []WorkflowBranch{{When: "test_first", Next: "qa_write_tests"}, {When: "dev_first", Next: "dev_implement"}}},
			{ID: "qa_write_tests", Type: "assign_role_task", Actor: "qa", Task: "先编写验收测试或测试用例", Next: "dev_implement"},
			{ID: "dev_implement", Type: "assign_role_task", Actor: "dev", Task: "根据需求与测试实现功能", Next: "qa_verify"},
			{ID: "qa_verify", Type: "verify_or_test", Actor: "qa", Task: "执行验证并反馈结果", Branches: []WorkflowBranch{{When: "passed", Next: "controller_finish"}, {When: "failed", Next: "dev_fix"}}},
			{ID: "dev_fix", Type: "assign_role_task", Actor: "dev", Task: "根据测试反馈修复问题", Next: "qa_verify"},
			{ID: "controller_finish", Type: "controller_task", Task: "决定是否合并与关闭流程", RequiresConfirmation: true, End: true},
		}
	default:
		return nil, fmt.Errorf("unsupported preset: %s", preset)
	}

	wf := &WorkflowTemplate{
		Version: 1,
		Name:    name,
		Roles: map[string]WorkflowRole{
			"cto": {Role: ctoRole},
			"dev": {Role: devRole},
			"qa":  {Role: qaRole},
		},
		Defaults: WorkflowDefaults{
			ExecutionMode:         executionMode,
			CreateWorkerIfMissing: true,
			ReuseWorker:           true,
		},
		Entry: "cto_breakdown",
		Nodes: nodes,
	}
	if errs := ValidateWorkflowTemplate(wf); len(errs) > 0 {
		return nil, fmt.Errorf("generated workflow is invalid: %s", strings.Join(errs, "; "))
	}
	return wf, nil
}

func LoadWorkflowRunState(path string) (*WorkflowRunState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow run state %s: %w", path, err)
	}
	var state WorkflowRunState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse workflow run state %s: %w", path, err)
	}
	return &state, nil
}

func (s *WorkflowRunState) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create workflow run-state directory: %w", err)
	}
	s.UpdatedAt = nowUTC()
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal workflow run state: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write workflow run state %s: %w", path, err)
	}
	return nil
}

func NewWorkflowRunState(workflowFile string, wf *WorkflowTemplate, runID string) *WorkflowRunState {
	now := nowUTC()
	state := &WorkflowRunState{
		Version:       1,
		RunID:         runID,
		WorkflowFile:  workflowFile,
		WorkflowName:  wf.Name,
		Status:        WorkflowStatusReady,
		CurrentNode:   wf.Entry,
		RoleWorkerMap: make(map[string]string),
		NodeStates:    make(map[string]WorkflowNodeState, len(wf.Nodes)),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	for _, node := range wf.Nodes {
		state.NodeStates[node.ID] = WorkflowNodeState{Status: WorkflowNodeStatusPending}
	}
	return state
}

func WorkflowRunID(workflowName string, t time.Time) string {
	return fmt.Sprintf("%s-%s", t.UTC().Format("20060102-150405"), slugify(workflowName))
}

func (wf *WorkflowTemplate) Node(nodeID string) (WorkflowNode, bool) {
	for _, node := range wf.Nodes {
		if node.ID == nodeID {
			return node, true
		}
	}
	return WorkflowNode{}, false
}

func (wf *WorkflowTemplate) ResolveNext(node WorkflowNode, outcome, explicitNext string) (string, error) {
	if explicitNext != "" {
		return explicitNext, nil
	}
	if outcome != "" && len(node.Branches) > 0 {
		for _, branch := range node.Branches {
			if branch.When == outcome {
				return branch.Next, nil
			}
		}
		return "", fmt.Errorf("no branch matches outcome %q for node %q", outcome, node.ID)
	}
	if node.Next != "" {
		return node.Next, nil
	}
	if node.End {
		return "", nil
	}
	return "", fmt.Errorf("node %q cannot resolve next step", node.ID)
}

func (s *WorkflowRunState) Start(wf *WorkflowTemplate, nodeID string, force bool) error {
	node, ok := wf.Node(nodeID)
	if !ok {
		return fmt.Errorf("node not found in workflow: %s", nodeID)
	}
	state := s.NodeStates[node.ID]
	if state.Status == WorkflowNodeStatusCompleted && !force {
		return fmt.Errorf("node already completed: %s", node.ID)
	}
	if state.StartedAt == "" {
		state.StartedAt = nowUTC()
	}
	state.Status = WorkflowNodeStatusRunning
	s.NodeStates[node.ID] = state
	s.Status = WorkflowStatusRunning
	s.CurrentNode = node.ID
	s.PendingConfirmation = ""
	s.BlockingReason = ""
	return nil
}

func (s *WorkflowRunState) Wait(wf *WorkflowTemplate, nodeID, reason string) error {
	node, ok := wf.Node(nodeID)
	if !ok {
		return fmt.Errorf("node not found in workflow: %s", nodeID)
	}
	state := s.NodeStates[node.ID]
	if state.StartedAt == "" {
		state.StartedAt = nowUTC()
	}
	state.Status = WorkflowNodeStatusWaiting
	state.Summary = reason
	s.NodeStates[node.ID] = state
	if node.RequiresConfirmation {
		s.Status = WorkflowStatusWaitingConfirmation
		s.PendingConfirmation = node.ID
	} else {
		s.Status = WorkflowStatusWaiting
		s.PendingConfirmation = ""
	}
	s.CurrentNode = node.ID
	s.BlockingReason = reason
	s.appendDecision(node.ID, "wait", "", "", reason)
	return nil
}

func (s *WorkflowRunState) Block(wf *WorkflowTemplate, nodeID, reason string) error {
	node, ok := wf.Node(nodeID)
	if !ok {
		return fmt.Errorf("node not found in workflow: %s", nodeID)
	}
	state := s.NodeStates[node.ID]
	if state.StartedAt == "" {
		state.StartedAt = nowUTC()
	}
	state.Status = WorkflowNodeStatusBlocked
	state.Summary = reason
	s.NodeStates[node.ID] = state
	s.Status = WorkflowStatusBlocked
	s.CurrentNode = node.ID
	s.PendingConfirmation = ""
	s.BlockingReason = reason
	s.appendDecision(node.ID, "block", "", "", reason)
	return nil
}

func (s *WorkflowRunState) Complete(wf *WorkflowTemplate, nodeID, outcome, explicitNext, workerID, summary, action string) (string, error) {
	node, ok := wf.Node(nodeID)
	if !ok {
		return "", fmt.Errorf("node not found in workflow: %s", nodeID)
	}
	state := s.NodeStates[node.ID]
	if state.Status == WorkflowNodeStatusCompleted {
		return "", fmt.Errorf("node already completed: %s", node.ID)
	}
	nextNode, err := wf.ResolveNext(node, outcome, explicitNext)
	if err != nil {
		return "", err
	}
	if workerID != "" && node.Actor != "" {
		if s.RoleWorkerMap == nil {
			s.RoleWorkerMap = make(map[string]string)
		}
		s.RoleWorkerMap[node.Actor] = workerID
	}
	if state.StartedAt == "" {
		state.StartedAt = nowUTC()
	}
	state.Status = WorkflowNodeStatusCompleted
	state.CompletedAt = nowUTC()
	state.Outcome = outcome
	state.Worker = workerID
	state.Summary = summary
	s.NodeStates[node.ID] = state
	s.PendingConfirmation = ""
	s.BlockingReason = ""
	s.CurrentNode = nextNode
	if nextNode == "" {
		s.Status = WorkflowStatusCompleted
	} else {
		s.Status = WorkflowStatusReady
	}
	s.appendDecision(node.ID, action, outcome, nextNode, summary)
	return nextNode, nil
}

func (s *WorkflowRunState) Confirm(wf *WorkflowTemplate, nodeID, outcome, explicitNext, workerID, summary string) (string, error) {
	node, ok := wf.Node(nodeID)
	if !ok {
		return "", fmt.Errorf("node not found in workflow: %s", nodeID)
	}
	if !node.RequiresConfirmation {
		return "", fmt.Errorf("node does not require confirmation: %s", node.ID)
	}
	return s.Complete(wf, nodeID, outcome, explicitNext, workerID, summary, "confirm")
}

func (s *WorkflowRunState) appendDecision(nodeID, action, outcome, nextNode, summary string) {
	s.DecisionLog = append(s.DecisionLog, WorkflowDecision{
		Timestamp: nowUTC(),
		Node:      nodeID,
		Action:    action,
		Outcome:   outcome,
		NextNode:  nextNode,
		Summary:   summary,
	})
}

func nowUTC() string {
	return time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)
}

func slugify(value string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(value) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteRune('-')
			prevDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
