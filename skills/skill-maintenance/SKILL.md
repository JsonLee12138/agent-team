---
name: skill-maintenance
description: >
  Skill cache maintenance skill for checking, updating, and cleaning installed skill artifacts.
  Use when the user asks to inspect or refresh skill cache state.
---

# skill-maintenance

## Audience

- human
- controller

## Triggers

- check skills
- update skills
- clean skills
- refresh skill cache

## CLI Binding

- `agent-team skill check`
- `agent-team skill update`
- `agent-team skill clean`

## Required Entry

- MUST read `.agent-team/rules/index.md` first.

## Expansion

- Load only the skill cache and dependency information needed for the maintenance action.

## Boundary

- This skill is for skill cache maintenance only.
- Do not route role-repo or rules sync work through this skill.
