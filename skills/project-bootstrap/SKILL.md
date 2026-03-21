---
name: project-bootstrap
description: >
  Project bootstrap and migration entry for initializing agent-team in a repository
  or migrating legacy project state.
---

# project-bootstrap

## Audience

- human
- controller

## Triggers

- init project
- migrate project

## CLI Binding

- `agent-team init`
- `agent-team migrate`

## Required Entry

- MUST read `.agents/rules/index.md` first.

## Expansion

- Load only the initialization-related config files and migration targets relevant to the requested action.

## Boundary

- This skill covers repository onboarding and migration.
- Do not use it for rules refresh or skill cache maintenance after bootstrap.
