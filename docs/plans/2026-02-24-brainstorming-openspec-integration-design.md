# Brainstorming + OpenSpec Integration Design

Date: 2026-02-24

## Summary

Integrate two capabilities into agent-team:

1. **Brainstorming** — Structured requirement exploration at the controller level (SKILL.md), executed before assigning work to roles
2. **OpenSpec** — Spec-driven change management replacing the file-based task system (`tasks/pending/` / `tasks/done/`)

## Architecture

```
User ──→ Controller AI (Claude) ──→ agent-team CLI ──→ Role AI (in worktree)
         │                          │
         │ brainstorming            │ assign (create change + proposal)
         │ (guided by SKILL.md)     │
         ↓                          ↓
     Design confirmed           Role receives notification
                                → /opsx:continue (specs, design, tasks)
                                → /opsx:apply (implement)
                                → /opsx:verify (validate)
```

### Layer Responsibilities

| Layer | Responsibility | Tool |
|-------|---------------|------|
| SKILL.md | Guide controller AI to brainstorm before assign | Text instructions |
| `agent-team create` | Create role + `openspec init` in worktree | Go CLI + exec OpenSpec |
| `agent-team assign` | Create OpenSpec change + write proposal.md + notify role | Go CLI + exec OpenSpec |
| `agent-team status` | Read OpenSpec change status per role | Go CLI + exec OpenSpec |
| Role prompt.md | Guide role to use `/opsx:continue` → `apply` → `verify` | Template text |

### Key Decisions

- Brainstorming runs at **controller level** (user/main AI), not inside roles. Roles are executors, not decision-makers.
- Brainstorming **replaces** OpenSpec's explore phase. Output goes directly into `proposal.md`.
- OpenSpec **completely replaces** the `tasks/` mechanism. No coexistence.
- OpenSpec CLI is called via `exec.Command` (approach A: minimal coupling).
- OpenSpec is **lazy-installed** at runtime if missing.

## Command Changes

### `create` — Add OpenSpec Init

Current:
1. Create worktree + branch
2. Create `agents/teams/<name>/config.yaml`
3. Create `agents/teams/<name>/prompt.md`
4. Create `tasks/pending/` and `tasks/done/`

New:
1. Create worktree + branch (unchanged)
2. Create `agents/teams/<name>/config.yaml` (unchanged)
3. Create `agents/teams/<name>/prompt.md` (template updated)
4. ~~Create `tasks/` directories~~ → instead:
   - Check if `openspec` CLI is available (`exec.LookPath("openspec")`)
   - If missing → run `npm install -g @fission-ai/openspec@latest`
   - Run `openspec init --tools claude,codex,opencode` in worktree directory

Resulting worktree structure:
```
.worktrees/<name>/
├── agents/teams/<name>/
│   ├── config.yaml
│   └── prompt.md
├── openspec/
│   ├── specs/
│   ├── changes/
│   └── config.yaml
└── CLAUDE.md              ← generated on open
```

### `assign` — Rewrite for OpenSpec

Current:
1. Create `tasks/pending/<timestamp>-<slug>.md`
2. Auto-open role if offline
3. Notify: `"New task assigned: <desc>\nPlease read: ..."`

New:
1. Accept brainstorming output via `--proposal` flag (file path or stdin with `-`)
2. In role worktree, create OpenSpec change:
   - Generate change name: `<timestamp>-<slug>` (reuse existing Slugify)
   - Create `openspec/changes/<change-name>/` directory
   - Create `.openspec.yaml` metadata file
   - Write brainstorming content to `proposal.md`
3. Auto-open role if offline (unchanged)
4. Notify: `"[New Change Assigned] <desc>\nChange: openspec/changes/<change-name>/\nProposal ready. Run /opsx:continue to proceed."`

Command signature:
```bash
# With proposal from brainstorming
agent-team assign <name> "<description>" --proposal <file-path>
agent-team assign <name> "<description>" --proposal -   # stdin

# Without proposal (simple scenario, creates empty proposal)
agent-team assign <name> "<description>"
```

### `status` — Read OpenSpec Status

Current:
```
Role            Status                   Pending Tasks
──────────────  ────────────────────────  ─────────────
frontend        ✓ running [p:50]          2
```

New:
```
Role            Status                   Changes
──────────────  ────────────────────────  ──────────────────────────
frontend        ✓ running [p:50]          2 active (1 implementing, 1 planning)
backend         ✗ offline                 1 active (1 ready)
```

Implementation: run `openspec status --json` in each role's worktree, parse JSON.

Status mapping:

| OpenSpec artifact state | Display as |
|------------------------|------------|
| proposal done, others blocked/ready | planning |
| tasks done, not applied | ready |
| applying | implementing |
| verify done | completed |

### Unchanged Commands

`open`, `open-all`, `reply`, `merge`, `delete` — logic unchanged.

## SKILL.md Changes

Add brainstorming flow guidance. When the user intends to assign new work:

1. Explore project context (check role prompt.md, existing specs)
2. Ask clarifying questions one at a time (prefer multiple choice)
3. Propose 2-3 approaches with trade-offs and recommendation
4. User confirms design
5. Write design to temp file
6. Execute `agent-team assign <name> "<desc>" --proposal <file>`

Rules:
- Brainstorming is **mandatory** before assign for new work
- Can be skipped when user explicitly says "just assign" or provides complete design
- One question at a time, YAGNI, explore alternatives

## Role prompt.md Template

```markdown
# Role: <name>

## Description
Describe this role's responsibilities here.

## Expertise
- List key areas of expertise

## Behavior
- How this role approaches tasks

## Workflow
When you receive a [New Change Assigned] message:
1. Read the proposal at the specified change path
2. Run /opsx:continue to create remaining artifacts (specs, design, tasks)
3. Run /opsx:apply to implement tasks
4. Run /opsx:verify to validate implementation
5. Commit your work regularly

## Communication Protocol
When you need clarification:
ask claude "<rolename>: <your question>"
Wait for reply (appears as [Main Controller Reply]).
Do NOT proceed on blocked tasks until reply arrives.
```

## Code Changes Summary

### Modify
- `cmd/create.go` — replace tasks/ creation with OpenSpec init
- `cmd/assign.go` — rewrite: create OpenSpec change + proposal
- `cmd/status.go` — read OpenSpec status instead of tasks/
- `internal/role.go` — update `PromptMDContent()`, `GenerateClaudeMD()`, remove tasks/ helpers
- `skills/agent-team/SKILL.md` — add brainstorming section, update assign docs
- Related test files

### Remove
- All `tasks/pending/` and `tasks/done/` related code

### Add
- `internal/openspec.go` — helper functions for OpenSpec CLI invocation (install check, init, status parsing)
- `--proposal` flag on assign command
