---
name: role-browser
description: >
  Read-only local role browsing skill for controller, worker, and human sessions.
  Use when the user wants to list available local roles without creating roles or managing repositories.
---

# role-browser

## Audience

- controller
- worker
- human

## Triggers

- role list
- browse roles
- view local roles

## CLI Binding

- `agent-team role list`

## Required Entry

- MUST read `.agent-team/rules/index.md` first.

## Expansion

- Load only the local role information required for the current read-only browsing request.

## Boundary

- This skill is only for browsing local roles.
- Do not absorb `role create` or `role-repo` actions.
