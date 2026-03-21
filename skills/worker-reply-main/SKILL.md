---
name: worker-reply-main
description: >
  Worker-first reporting entry for sending completion, blocker, or decision-needed messages back to main
  through `agent-team reply-main`.
---

# worker-reply-main

## Audience

- worker

## Triggers

- reply main
- report completion
- blocked
- request decision

## CLI Binding

- `agent-team reply-main`

## Required Entry

- MUST read `worker.yaml` first.

## Expansion

- Prepare only the minimum task summary needed for a factual reply to main.

## Boundary

- This skill is only for worker -> main communication.
- Do not use it for controller -> worker replies; that belongs to `worker-dispatch`.
- Do not turn this skill into a recovery or planning surface.
