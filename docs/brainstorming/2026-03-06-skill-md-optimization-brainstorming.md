# Brainstorming: SKILL.md Optimization

- **Role**: General Strategist
- **Date**: 2026-03-06
- **Status**: Approved

## Problem Statement

`skills/agent-team/SKILL.md` has fallen behind the actual CLI capabilities. Multiple command groups (`catalog`, `task`, `role-repo check/update/remove`), AI behavior directives (replacing hooks), and key workflows (role discovery) are undocumented. The file structure is flat and hard to navigate as features grow.

## Goals

1. **Complete coverage** — Document all CLI features the controller AI needs
2. **Streamline** — Remove verbose inline details, delegate to references
3. **Restructure** — Organize by lifecycle phases for intuitive reading order
4. **Replace hooks with AI behavior directives** — More universal and stable
5. **Clean up OpenSpec remnants** — Remove all stale OpenSpec references
6. **Add Role Discovery Flow** — Guide controller when a role doesn't exist locally

## Constraints & Assumptions

- Primary audience: Controller AI (with worker behavior understanding for context)
- SKILL.md should be self-sufficient for core workflows; detailed params go to references
- All destructive operations (delete, merge) require explicit user confirmation
- No hooks approach — all automation is AI-driven via behavior directives in SKILL.md
- OpenSpec has been replaced by `.tasks/` system; residual references must be cleaned

## Design: Approach C — Lifecycle Phases + Inline AI Behavior Directives

### Why Approach C

Compared to alternatives:
- **Approach A** (separate AI behavior chapter): Causes duplication with lifecycle sections
- **Approach B** (heavy references split): AI may not proactively read references, causing missed directives
- **Approach C** (inline): AI behavior directives live alongside the operations they govern — no duplication, no missed context

### Structure Overview

```
1. Frontmatter (updated triggers)
2. Overview + Core Concepts (Role/Worker model)
3. Install → references/install-upgrade.md
4. Phase 1: Initialization (init) + [AI: prompt init when missing]
5. Phase 2: Role Preparation (create, team, discovery flow, role-repo full suite)
6. Phase 3: Worker Lifecycle (create, open, assign, status, merge, close, delete)
7. Phase 4: Task & Monitoring (task commands reference, worker TDD cycle)
8. Phase 5: Communication (reply, reply-main)
9. Catalog (trigger-based, query commands only, admin → --help)
10. Backend Selection
11. Migration
```

### Section Details

#### Frontmatter

```yaml
---
name: agent-team
description: >
  AI team role and worker manager for multi-agent development workflows.
  Uses Role (skill package) + Worker (instance) model with git worktrees.
  Triggers: /agent-team, create a team, batch create roles, create a role,
  create a worker, open worker session, assign task, show team status,
  merge worker branch, search catalog, find role, install role,
  check task status, role-hub.
---
```

Changes from current:
- Triggers expanded with `search catalog`, `find role`, `install role`, `check task status`, `role-hub`
- Install section moved to one-line reference link

#### Phase 1: Initialization

- `agent-team init [--global-only] [--skip-detect]` — concise description
- **AI Behavior**: Detect missing `.agents/teams/`, prompt user to run init. Do NOT auto-run.

#### Phase 2: Role Preparation

Four sub-sections:

1. **Creating a Role** — `/role-creator` skill with `--target-dir .agents/teams`
2. **Creating a Team (Batch)** — Parse prompt → normalize kebab-case → confirm → create. Stop after creation, do NOT create workers.
3. **Role Discovery Flow (NEW)** — When role not found locally:
   - Ask user: "Create from scratch?" or "Search Role Hub?"
   - If search: `agent-team catalog search <query>` → show results → `agent-team role-repo add <source> --role <name>`
   - If no match or declined: fall back to create from scratch
4. **Role Repo Management** — All 6 subcommands: `find`, `add`, `list`, `check`, `update`, `remove`
5. **Listing Local Roles** — `agent-team role list`

#### Phase 3: Worker Lifecycle

- **Create**: `agent-team worker create <role-name> [provider]` — creates worktree at `.worktrees/<worker-id>/` and automatically opens session window
- **Open**: `agent-team worker open <worker-id>` — reopen existing worker session. Platform reference link.
  - **AI Behavior (Session Start)**: Ensure worker loads role context and is aware of pending tasks
- **Assign**: `agent-team worker assign <worker-id> "<description>"` — creates task change at `.tasks/changes/`
  - Pre-assign requirements:
    1. Brainstorming completed → references/brainstorming.md
    2. Readiness Gate passed → references/readiness.md
    3. **Sync worktree with main**: If worker exists with no uncommitted changes, rebase/merge latest main. If uncommitted changes exist, ask user first.
- **Status**: `agent-team worker status`
- **Merge**: `agent-team worker merge <worker-id>` — merges with `--no-ff`
- **Close Session**: Close terminal without deleting worktree/branch. Can reopen later.
- **Delete**: `agent-team worker delete <worker-id>` — irreversible.

**AI Behavior — Worker Cleanup Rules:**
- After task completion notification: do NOT auto-merge. Ask user: "Review code, merge, or assign next task?"
- After merge: do NOT auto-delete. Ask user: "Keep worker or delete?"
- NEVER run `worker delete` without explicit user approval
- Closing session does NOT require confirmation

**AI Behavior — Idle Worker Scheduling:**
- Inform user when worker idle + pending tasks. Do NOT auto-assign.

#### Phase 4: Task & Monitoring

- Task commands listed as reference (list, show, done, verify, archive)
- Worker TDD cycle summary (read → test → implement → verify → notify)
- Worker notification formats (completed/failed/skipped/blocked)
- **AI Behavior (Task Completion)**: Check verify result, surface to user, ask how to proceed. Do NOT auto-merge.

#### Phase 5: Communication

- `agent-team reply <worker-id> "<answer>"` — controller → worker
- `agent-team reply-main "<message>"` — worker → controller
- **AI Behavior**: Surface worker questions to user immediately. Do NOT answer on behalf of user. Batch-present multiple blocked workers.

#### Catalog (Role Hub)

- Query commands only: `search`, `show`, `list`, `repo`, `stats`
- Admin commands (discover, normalize, serve) → `--help`
- Links back to Phase 2 Role Discovery Flow

#### Backend Selection

- Default: WezTerm. `AGENT_TEAM_BACKEND=tmux` for tmux.

#### Migration

- `agent-team migrate` — `agents/` → `.agents/`

### OpenSpec Cleanup (Code Changes)

| File | Change |
|------|--------|
| `cmd/worker_assign.go:21` | Short description: `"Create an OpenSpec change..."` → `"Create a task change and notify the worker session"` |
| `internal/role_test.go:229-240` | Evaluate `OPENSPEC:START/END` marker test; remove if no longer needed |
| SKILL.md full text | Ensure zero OpenSpec references |

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| SKILL.md becomes too long after adding all features | Heavy use of references; only core flow + AI behavior inline |
| AI ignores `> AI Behavior` blocks | Use consistent formatting; test with actual AI sessions |
| Role Discovery Flow adds latency to role creation | Search is optional — user can always choose "create from scratch" |
| Worktree sync before assign may cause conflicts | Ask user when uncommitted changes detected; do not force |

## Validation Strategy

1. Diff current SKILL.md vs new version — verify no unintended deletions
2. Grep new SKILL.md for `openspec` — must return zero results
3. Verify all CLI subcommands are covered (cross-reference `agent-team --help` recursive)
4. Test AI behavior directives: simulate controller session and verify it follows the rules
5. Build and test OpenSpec code cleanup (`go test ./...`)

## Open Questions

None — all sections approved.
