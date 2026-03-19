# Workflow Execution Contract

## Controller Responsibilities

- Validate the workflow template before every new run.
- Prepare or reuse workers only when the workflow actually needs them.
- Persist state before and after every meaningful transition.
- Surface blocked decisions to the user instead of guessing.
- Keep `agent-team` as a fixed backend and compose public commands only.
- Route context-management triggers through `strategic-compact` instead of embedding workflow-local compact policy.

## Node Execution Matrix

### `controller_task`

- Execute directly in the controller session.
- If `requires_confirmation: true`, pause and use `agent-team workflow state confirm`.
- Otherwise complete the node and auto-advance through `next` or `branches`.

### `assign_role_task`

- Resolve `actor` to a role alias, then to a worker id.
- Create the worker if allowed by workflow defaults and no worker exists yet.
- Open the worker session before assignment.
- Send the task with `agent-team worker assign`.
- Mark the run state as `waiting` until the worker reports back.

### `wait_for_completion`

- Use when dispatch and polling are separated into two nodes.
- Inspect worker status and controller messages.
- Convert the observed result into a workflow outcome, then complete or block.

### `decision`

- Use for controller-visible branch selection.
- Record the chosen `when` label in `decision_log`.
- Never invent an outcome that is not present in the node definition.

### `handoff`

- Use when ownership moves from one actor to another.
- Record the handoff summary in run state.
- Treat the target node as a separate assignment or controller node.

### `verify_or_test`

- Run local verification or delegate QA verification.
- Convert results into explicit outcomes such as `passed` or `failed`.
- On `failed`, branch back to a remediation node such as `dev_fix`.

### `merge`

- Ask for explicit controller approval before merging.
- Run `agent-team worker merge <worker-id>` only after approval.
- If merge fails, mark the run `blocked` with the error summary.

## Pause Conditions

Pause immediately when any of these occur:

- a node requires controller confirmation
- worker feedback is ambiguous
- a referenced role or worker cannot be prepared
- a command fails
- state cannot be written safely

Before a pause caused by `wait`, `block`, or resumed control after a restart, check whether `strategic-compact` should create a recovery anchor for main.

When paused:

1. Persist the current node and blocking reason.
2. Keep completed nodes immutable unless the controller requests a retry.
3. Resume from the same run-state file instead of re-initializing the run.

## Recommended Command Sequence

### New Run

```bash
agent-team workflow validate <workflow-file>
agent-team workflow state init <workflow-file>
agent-team workflow state show <state-file>
```

### Dispatch To Worker

```bash
# evaluate strategic-compact if this starts a new controller phase
agent-team workflow state start <state-file> --node <node-id>
agent-team worker assign <worker-id> "<task>"
# evaluate strategic-compact before the controller enters waiting state if context should be checkpointed
agent-team workflow state wait <state-file> --node <node-id> --reason "waiting for worker reply"
```

### Resolve Worker Reply

```bash
# evaluate strategic-compact before large worker output or review material is read
agent-team workflow state complete <state-file> --node <node-id> --summary "<reply summary>"
agent-team workflow state show <state-file>
```

### Resolve Confirmation

```bash
agent-team workflow state confirm <state-file> --node <node-id> --outcome dev_first
agent-team workflow state show <state-file>
```
