---
name: task-orchestrator
description: >
  Scenario-first task lifecycle entry for controller and human sessions.
  Use when the user asks to create, assign, complete, archive, or inspect task flow through `agent-team task` commands.
---

# task-orchestrator

## Audience

- controller
- human

## Triggers

- create task
- assign task
- complete task
- archive task
- task flow

## CLI Binding

- `agent-team task create`
- `agent-team task list`
- `agent-team task show`
- `agent-team task assign`
- `agent-team task done`
- `agent-team task archive`

## Required Entry

- MUST read `.agents/rules/index.md` first.

## Expansion

- Load only the task artifacts and rule files required by the current lifecycle action.

## Boundary

- This is the write-capable task lifecycle entry.
- Do not use this skill for worker-only recovery or worker -> main reporting.
- If the user only wants read-only task status with no lifecycle mutation intent, prefer `task-inspector`.
