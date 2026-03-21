---
name: role-repo-manager
description: >
  Role source management skill for finding, adding, listing, checking, updating, and removing role repositories.
  Use when the user is managing where roles come from rather than browsing the catalog.
---

# role-repo-manager

## Audience

- human
- controller

## Triggers

- search role repo
- add role repo
- update role repo
- remove role repo

## CLI Binding

- `agent-team role-repo find`
- `agent-team role-repo add`
- `agent-team role-repo list`
- `agent-team role-repo check`
- `agent-team role-repo update`
- `agent-team role-repo remove`

## Required Entry

- MUST read `.agent-team/rules/index.md` first.

## Expansion

- Load only the target role repo results and the selected role information needed for the requested action.

## Boundary

- This skill manages role sources.
- Do not use it for read-only catalog browsing when no source mutation is intended; that belongs to `catalog-browser`.
