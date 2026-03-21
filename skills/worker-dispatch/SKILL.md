---
name: worker-dispatch
description: >
  Controller-side worker dispatch entry for opening workers, checking targeted worker status,
  and replying to workers with existing `agent-team worker` and `agent-team reply` commands.
---

# worker-dispatch

## Audience

- controller
- human

## Triggers

- open worker
- dispatch worker
- reply to worker
- inspect current worker before replying

## CLI Binding

- `agent-team worker open`
- `agent-team worker status`
- `agent-team reply`

## Required Entry

- MUST read `.agent-team/rules/index.md` first.

## Expansion

- Load the relevant worker config and referenced task artifact only when required by the dispatch action.

## Boundary

- This is the controller -> worker dispatch surface.
- Do not use this skill to recover a worker's own context; that belongs to `worker-recovery`.
- For read-only fleet inspection with no dispatch or reply intent, prefer `worker-inspector`.
