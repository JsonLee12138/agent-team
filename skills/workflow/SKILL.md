---
name: workflow
description: >
  Manage governance-only workflow plans through `agent-team workflow plan`.
  Use when the user asks to generate, approve, activate, or close workflow plans.
---

# Workflow (Governance-Only)

## Overview

This skill is governance-only.

Use `agent-team workflow plan` as the single workflow surface:

- `generate`
- `approve`
- `activate`
- `close`

Legacy execution-surface commands (`workflow create|validate|state`) are retired.

## Hard Rules

- Act as the main controller only.
- Do not reintroduce execution-surface behavior in this skill.
- Before plan actions, ensure required governance inputs are explicit (`plan-id`, `task-id`, `owner`, refs/evidence when needed).
- On gate blockers, report blocker code/message directly; do not bypass or auto-downgrade.
- Keep CLI output minimal and factual (`plan_id`, `status`, blocker summary).

## Action Routing

### Generate Plan

Use when the user asks to create a governance plan proposal.

```bash
agent-team workflow plan generate \
  --plan-id <plan-id> \
  --task-id <task-id> \
  --owner <owner> \
  [--module workflow] \
  [--ref <id>]... \
  [--evidence <ref>]... \
  [--reason <text>]... \
  [--use-archived]
```

Expected result:

- `status=proposed`

### Approve Plan

Use when owner signoff is requested.

```bash
agent-team workflow plan approve \
  --plan-id <plan-id> \
  --actor <owner-id>
```

Expected result:

- `status=approved`

### Activate Plan

Use when an approved plan should become active.

```bash
agent-team workflow plan activate \
  --plan-id <plan-id>
```

Expected result:

- `status=active`

### Close Plan

Use when an active plan should be closed.

```bash
agent-team workflow plan close \
  --plan-id <plan-id>
```

Expected result:

- `status=closed`

## Failure Handling

- Gate blocker: return as hard failure and surface blocker details.
- Invalid state transition: return command error directly.
- Owner mismatch on approve: fail without fallback approver.

## Commands

- `agent-team workflow plan generate`
- `agent-team workflow plan approve`
- `agent-team workflow plan activate`
- `agent-team workflow plan close`
