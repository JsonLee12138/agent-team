# Workflow Plan Schema (Governance-Only)

## Workflow Plan Object

The governance plan persisted by `workflow plan` commands contains fields equivalent to:

```yaml
id: plan-123
task_id: task-1
owner: owner-1
status: proposed # proposed|approved|active|closed
created_at: "2026-03-20T10:00:00Z"
approved_at: "2026-03-20T10:10:00Z"
activated_at: "2026-03-20T10:20:00Z"
closed_at: "2026-03-20T10:30:00Z"
input_refs:
  - req:task-1
reasoning:
  - "why this plan exists"
```

## Status Rules

- Initial status from generate: `proposed`
- Owner approval transitions to: `approved`
- Activation transitions to: `active`
- Closing transitions to: `closed`

Invalid transitions must fail with an explicit error.

## Command Surface

Supported commands:

- `agent-team workflow plan generate`
- `agent-team workflow plan approve`
- `agent-team workflow plan activate`
- `agent-team workflow plan close`

Retired commands are out of scope for this skill.
