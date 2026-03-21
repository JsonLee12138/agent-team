---
name: workflow-orchestrator
description: >
  Governance-only workflow plan entry for controller and human sessions.
  Use when the user asks to generate, approve, activate, or close workflow plans through `agent-team workflow plan`.
---

# workflow-orchestrator

## Audience

- controller
- human

## Triggers

- workflow plan
- approve plan
- activate plan
- close plan

## CLI Binding

- `agent-team workflow plan generate`
- `agent-team workflow plan approve`
- `agent-team workflow plan activate`
- `agent-team workflow plan close`

## Required Entry

- MUST read `.agent-team/rules/index.md` first.

## Expansion

- Load only the workflow governance artifacts required for the requested plan action.

## Boundary

- This skill is governance-only.
- Do not promise execution surfaces that do not exist.
- Do not use this skill for worker assignment, worker recovery, or worker status inspection.
