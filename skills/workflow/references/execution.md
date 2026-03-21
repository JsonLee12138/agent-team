# Workflow Governance Execution Contract

## Controller Responsibilities

- Use only governance plan lifecycle commands.
- Keep actions on the state path: `proposed -> approved -> active -> closed`.
- Surface gate blockers and transition errors immediately.
- Do not invoke retired execution-surface commands.

## Lifecycle Command Matrix

### `generate`

Purpose: create a proposed workflow governance plan.

```bash
agent-team workflow plan generate \
  --plan-id <plan-id> \
  --task-id <task-id> \
  --owner <owner>
```

### `approve`

Purpose: owner signoff.

```bash
agent-team workflow plan approve \
  --plan-id <plan-id> \
  --actor <owner>
```

### `activate`

Purpose: activate an approved plan.

```bash
agent-team workflow plan activate \
  --plan-id <plan-id>
```

### `close`

Purpose: close an active plan.

```bash
agent-team workflow plan close \
  --plan-id <plan-id>
```

## Guardrails

- Approve requires owner actor.
- Activate requires current status `approved`.
- Close requires current status `active`.
- Any blocker/invalid transition keeps the plan in its prior status.
