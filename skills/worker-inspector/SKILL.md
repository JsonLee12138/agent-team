---
name: worker-inspector
description: >
  Read-only worker inspection skill for controller and human sessions.
  Use when the user wants to view worker status without opening a worker or sending a reply.
---

# worker-inspector

## Audience

- controller
- human

## Triggers

- worker status
- inspect worker
- list workers

## CLI Binding

- `agent-team worker status`

## Required Entry

- MUST read `.agent-team/rules/index.md` first.

## Expansion

- Load only the worker status information needed for the read-only inspection request.

## Boundary

- This skill is read-only.
- If the prompt implies open, dispatch, or reply behavior, route to `worker-dispatch` instead.
