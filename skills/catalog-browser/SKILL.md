---
name: catalog-browser
description: >
  Read-only catalog browsing skill for searching and viewing role catalog information.
  Use when the user wants to list, search, inspect, or review catalog repositories and statistics.
---

# catalog-browser

## Audience

- human
- controller

## Triggers

- search catalog
- browse roles
- show catalog role
- catalog stats

## CLI Binding

- `agent-team catalog search`
- `agent-team catalog show`
- `agent-team catalog list`
- `agent-team catalog repo`
- `agent-team catalog stats`

## Required Entry

- MUST read `.agent-team/rules/index.md` first.

## Expansion

- Load only the catalog results needed for the current read-only browsing request.

## Boundary

- This skill exposes only the browsing subset of catalog commands.
- Do not absorb role source installation or update flows; those belong to `role-repo-manager`.
- Do not expose `normalize`, `discover`, or `serve` through this skill.
