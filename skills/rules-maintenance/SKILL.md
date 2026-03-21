---
name: rules-maintenance
description: >
  Rule refresh skill for syncing built-in rule templates and generated rule artifacts.
  Use when the user asks to sync rules or fix rule drift.
---

# rules-maintenance

## Audience

- human
- controller

## Triggers

- sync rules
- update rules
- rule drift

## CLI Binding

- `agent-team rules sync`

## Required Entry

- MUST read `.agent-team/rules/index.md` first.

## Expansion

- Load only the rule files implicated by the requested sync or drift investigation.

## Boundary

- This skill is for rule maintenance only.
- Do not absorb skill cache operations; those belong to `skill-maintenance`.
