# State Collector

## Goal

Collect the minimum bounded state needed to make `/compact` useful without paying the cost of a broad history reread.

The collector is intentionally small. It should prefer already-structured state sources over reconstructing the whole conversation.

## Minimum State Schema

```yaml
current_goal: string
current_phase: string
task_status: string | object
workflow_status: string | object
constraints:
  - string
recent_decisions:
  - string
completed_items:
  - string
pending_items:
  - string
next_step: string
risks_or_blockers:
  - string
verification_state: string
```

## Source Priority

### 1. Explicit controller intent

Use the current approved plan, current assignment objective, or workflow node objective first.

Maps best to:
- `current_goal`
- `current_phase`
- `next_step`
- `constraints`

### 2. Workflow run state

Prefer `WorkflowRunState` as the main structured source for controller context.

Relevant fields from `internal/workflow.go`:
- `Status`
- `CurrentNode`
- `BlockingReason`
- `NodeStates`
- `DecisionLog`
- `RoleWorkerMap`

Use them to derive:
- `workflow_status`
- `current_phase`
- `recent_decisions`
- `completed_items`
- `pending_items`
- `risks_or_blockers`

### 3. Task/change lifecycle state

Use task/change status only as supporting context, not as the primary orchestrator state.

Relevant signals:
- change status transitions via `ApplyChangeTransition()`
- verify/done/archive progression
- worker completion chain outcome

Use them to derive:
- `task_status`
- `verification_state`
- `completed_items`
- `pending_items`

### 4. Recent controller decisions

Prefer the latest bounded decisions over full transcript replay.

Examples:
- selected workflow branch
- chosen next actor
- decision to wait/block/resume
- decision to review before merge

## Bounded Read Rules

- Read the smallest structured file or command output that can answer the schema.
- Prefer current workflow node plus recent decision log over full run history.
- Prefer current change status plus verification result over replaying all worker discussion.
- Do not read large diffs or large logs just to prepare compact for a large read.

## Collection Targets By Trigger

### `light`

Collect only:
- `current_goal`
- `current_phase`
- `completed_items`
- `next_step`
- `risks_or_blockers`

### `standard`

Collect the full minimum schema.

### `deep`

Only if escalation conditions are met. Deep may inspect wider decision history, but v1 does not require an implementation path.

## Output Quality Rule

If the minimum schema cannot be filled confidently, emit the missing field explicitly instead of inventing detail.

Example:

```yaml
verification_state: unknown
risks_or_blockers:
  - verification result not yet read
```
