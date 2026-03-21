---
name: context-cleanup
description: >
  Session context cleanup and file re-anchoring strategy for controller and worker sessions.
  Use when the session is drifting, phases change, or resumed work needs index-first recovery from files.
---

# context-cleanup

## Audience

- controller
- worker

## Triggers

- clean context
- session drift
- re-anchor context
- resume after pause
- phase transition

## CLI Binding

- `agent-team context-cleanup`
- This skill is NOT equivalent to `compact` and is not a wrapper around `/compact`.

## Required Entry

- controller/main: MUST read `.agents/rules/index.md` first.
- worker: MUST read `worker.yaml` first.

## Expansion

- controller/main: read only the rule files matched from `.agents/rules/index.md`, then the current workflow/task artifacts.
- worker: read `task.yaml` after `worker.yaml`, then `context.md` and referenced materials only when needed.

## Hard Rules

- Clean session context, not file contents.
- Never skip the required entry file.
- Never default to scanning all rule bodies or all task context files.
- Never describe this skill as a synonym for `/compact` or context compression.

## Boundary

- Use this skill to stabilize context and re-anchor from files.
- Use `worker-recovery` for routine worker task resume when the problem is simply recovering the current assignment.
