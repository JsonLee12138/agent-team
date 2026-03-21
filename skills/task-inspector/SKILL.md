---
name: task-inspector
description: >
  Read-only task inspection skill for controller, worker, and human sessions.
  Use when the user wants to view task status, inspect a task, or list tasks without changing lifecycle state.
---

# task-inspector

## Audience

- controller
- worker
- human

## Triggers

- inspect task
- task status
- list tasks
- show task

## CLI Binding

- `agent-team task list`
- `agent-team task show`

## Required Entry

- controller/human: MUST read `.agent-team/rules/index.md` first.
- worker: MUST read `worker.yaml` first.

## Expansion

- Load only the task artifact or task lookup needed for the read-only question.

## Boundary

- This skill is read-only.
- If the prompt includes create, assign, done, or archive intent, route to `task-orchestrator` instead.
