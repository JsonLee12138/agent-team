---
name: strategic-compact
description: >
  Main-first, token-first context management orchestration for controller sessions.
  Use when a main/controller session hits a compact trigger such as phase transition,
  context pressure, pre-large-read, manual recovery, or resume-after-pause. Reuse
  `agent-team compact` as the execution primitive instead of inventing a new compact path.
---

# strategic-compact

## Overview

Use this skill to keep the main/controller session as the primary context object.
Prefer compacting main before large reads, phase shifts, and resumed work. Keep
workers short-lived by default: finish -> main review/merge -> cleanup.

This skill is a strategy layer, not a new transport layer.

- It decides **when** a compact pass is required.
- It decides **how much** state to recover before compacting.
- It reuses `agent-team compact` or an existing recorded Claude pane as the execution primitive.
- It does not redefine pane discovery, provider transport, or workflow runtime behavior.

## Hard Rules

- Treat the main/controller session as the default compact target.
- Do not expand worker lifetime just to preserve context.
- Do not manually rebuild `/compact` transport logic inside rules or workflow docs.
- Use the smallest recovery payload that still preserves Goal / Phase / Constraints / Done / Next / Risks.
- Prefer `light` or `standard` strategy. `deep` is an escalation contract only in v1.
- Only allow worker compact as an exception path, not as the standard completion flow.

## Trigger Types

Use one of these trigger labels when invoking or reasoning about this skill:

- `manual`
- `phase-transition`
- `context-pressure`
- `pre-large-read`
- `resume-after-pause`

Trigger details: [references/triggers.md](references/triggers.md)

## State Collection Contract

Before compacting, collect only the minimum state needed to restore control:

- `current_goal`
- `current_phase`
- `task_status`
- `workflow_status`
- `constraints`
- `recent_decisions`
- `completed_items`
- `pending_items`
- `next_step`
- `risks_or_blockers`
- `verification_state`

Collector scope and source mapping: [references/state-collector.md](references/state-collector.md)

## Strategy Selection

### `light`

Use for quick phase shifts and pre-large-read checkpoints.

Typical cases:
- before reading a large diff, test log, or worker reply
- before `wait`, `start`, `complete`, or `confirm` changes workflow state
- before short resumed work when the current goal is still clear

### `standard`

Use for normal main-session maintenance when context is drifting but still recoverable with a bounded read.

Typical cases:
- several worker replies or controller decisions have accumulated
- the main session needs a durable recovery anchor before the next review / dispatch cycle
- resumed work needs a fresh controller summary from workflow and task state

### `deep`

Do not use as the default path in v1.

Only define escalation conditions now. See [references/strategies.md](references/strategies.md).

## Recovery Contract

Every compact preparation should be able to restore these anchors after the large read or resumed session:

- Goal
- Phase
- Constraints
- Done
- Next
- Risks

Detailed recovery format: [references/recovery-contract.md](references/recovery-contract.md)

## Main-First Policy

Main-first policy and worker exception rules: [references/main-first.md](references/main-first.md)

## Integration Contract

Use this skill from controller-side skills and rules only:

- `workflow` calls it around phase transitions, waits, resumes, and large reads.
- `agent-team` calls it when main is about to review worker output or start a new dispatch/review cycle.
- `.agents/rules/context-management.md` defines when it becomes mandatory.

Integration details: [references/integration.md](references/integration.md)

## Execution Primitive

Use existing compact delivery only:

```bash
agent-team compact --to main
agent-team compact --to main --message "Goal: ... Next: ..."
```

If the target is already a known Claude pane, injecting `/compact` into that recorded pane is equivalent.
Do not create a second compact transport contract inside this skill.
