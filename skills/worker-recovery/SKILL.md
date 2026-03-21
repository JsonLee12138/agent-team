---
name: worker-recovery
description: >
  Worker-first recovery entry for resuming the current assignment from file artifacts.
  Use when a worker needs to recover task context, continue work, or re-anchor after a pause.
---

# worker-recovery

## Audience

- worker

## Triggers

- resume work
- recover task
- continue current assignment

## CLI Binding

- Read worker artifacts directly.
- Use `agent-team task show` only when task details must be refreshed from the CLI.

## Required Entry

- MUST read `worker.yaml` first.

## Expansion

- Recovery order is fixed: `worker.yaml` -> `task.yaml` -> `context.md` -> referenced materials only if needed.

## Boundary

- This is the standard worker recovery path.
- Do not use controller assumptions or `.agent-team/rules/index.md` as the first entry here.
- Do not use this skill for worker -> main reporting; that belongs to `worker-reply-main`.
