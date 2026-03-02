---
name: agent-team
description: >
  AI team role and worker manager for multi-agent development workflows.
  Uses Role (skill package) + Worker (instance) model with git worktrees.
  Use when the user wants to create a team, create roles, manage workers,
  assign tasks, check team status, or merge worker branches.
  Triggers on /agent-team commands, "create a team", "batch create roles",
  "create a role", "create a worker", "open worker session",
  "assign task", "show team status", "merge worker branch".
---

# agent-team

Manages AI team workflows using a **Role + Worker** dual-layer model with git worktrees and terminal multiplexer tabs.

- **Role** = a skill package definition (SKILL.md + system.md + role.yaml) stored in `agents/teams/<role-name>/`
- **Worker** = a role instance running in an isolated worktree, identified by `<role-name>-<3-digit-number>`

For directory layout details, see [references/details.md](references/details.md).

## Install

macOS/Homebrew:

```bash
brew tap JsonLee12138/agent-team && brew install agent-team
```

Source (Go 1.24+):

```bash
go install github.com/JsonLee12138/agent-team@latest
```

## Upgrade

### Skill

Before running, MUST confirm with the user:

1. **Platform**: which agent platform(s) — see [references/platforms.md](references/platforms.md)
2. **Scope**: project-level or global

```bash
# Project-level
npx skills add JsonLee12138/agent-team -a <platform> -y

# Global
npx skills add JsonLee12138/agent-team -a <platform> -y -g
```

### CLI

```bash
# Homebrew
brew update && brew upgrade agent-team

# Source
go install github.com/JsonLee12138/agent-team@latest
```

## Role Management (AI Workflow)

Roles are created and managed by AI using the **role-creator** skill. The CLI does not handle role creation.

### Creating a Role

1. Use the `/role-creator` skill with `--target-dir agents/teams`
2. If a role already exists in global `~/.claude/skills/`, prompt user to copy it to `agents/teams/`
3. Result: `agents/teams/<role-name>/` with SKILL.md, system.md, references/role.yaml

### Creating a Team (Batch Role Creation from Prompt)

Use this workflow when the user describes a team in natural language, for example:

- "Create a frontend developer, a QA engineer, and a frontend architect role."
- "I need frontend dev + testing + frontend architecture roles."

Flow:

1. Parse the prompt into responsibility units (one role per responsibility).
2. Normalize each role name to kebab-case, such as:
   - frontend developer -> `frontend-dev`
   - QA engineer -> `qa-engineer`
   - frontend architect -> `frontend-architect`
3. Present the draft role list and ask for one confirmation before execution.
4. For each approved role, run the full **Creating a Role** workflow with `/role-creator --target-dir agents/teams`.
5. Return a per-role summary: `created`, `already exists`, or `failed` (with reason).
6. Stop after role creation. Do NOT create workers in this flow.

Rules:

- Team creation in this skill is role-only. It MUST NOT auto-run `agent-team worker create`, `worker open`, or `worker assign`.
- If a role already exists, do not overwrite it by default. Mark it as `already exists`.
- A single role failure does not cancel the whole batch. Continue and report final results.

### Listing Roles

```bash
agent-team role list
```

Shows all available roles in `agents/teams/`.

## Worker Management (CLI Commands)

### Create a worker

```bash
agent-team worker create <role-name>
```

Creates worktree `.worktrees/<worker-id>/` with branch `team/<worker-id>`, generates config, and initializes `.tasks/` directory.

### Open a worker session

```bash
agent-team worker open <worker-id> [claude|codex|opencode] [--model <model>] [--new-window]
```

Copies role + dependency skills into the worktree, generates CLAUDE.md/AGENTS.md from role's system.md, and opens a terminal tab with the chosen AI provider.

### Assign a change

```bash
agent-team worker assign <worker-id> "<description>" [provider] [--proposal <file>] [--design <file>] [--verify-cmd <cmd>] [--new-window]
```

Creates a task change at `.tasks/changes/<timestamp>-<slug>/` with:
- `--proposal` → `proposal.md` (work requirements)
- `--design` → `design.md` (architecture decisions)
- `--verify-cmd` → stored in `change.yaml` for worker verification

The CLI auto-opens the worker session if not running and sends a `[New Change Assigned]` notification.

**Before running this command**, the controller MUST execute the Assign Readiness Gate.

### Assign Readiness Gate (Required)

Before running `agent-team worker assign ...`, the controller MUST execute a readiness handshake.

- MUST send `AGENT_TEAM_PING <worker-id> <attempt>` and wait for matching `AGENT_TEAM_PONG <worker-id> <attempt>`
- MUST use: first wait `5s`, retry interval `5s`, maximum `3` attempts
- MUST inspect worker window on timeout before retrying
- MUST send provider open command when `codex`, `claude`, or `opencode` workspace is not open
- MUST fail fast after attempt 3 with structured error output and stop assign

For protocol details, diagnostics checklist, retry state flow, and error templates, see [references/readiness.md](references/readiness.md).

### Check status

```bash
agent-team worker status
```

Shows all workers, their roles, session status, and active changes.

### Merge completed work

```bash
agent-team worker merge <worker-id>
```

Merges `team/<worker-id>` into the current branch with `--no-ff`.
After merging, do NOT automatically delete the worker.

### Delete a worker

```bash
agent-team worker delete <worker-id>
```

Closes the running session, removes the worktree, deletes the branch, and cleans up `agents/workers/<worker-id>/`.

## Communication

### Reply to a worker

```bash
agent-team reply <worker-id> "<answer>"
```

### Reply to main controller (used by workers)

```bash
agent-team reply-main "<message>"
```

Messages appear in the controller's terminal as `[Worker: <worker-id>] <message>`.

For the full bidirectional communication protocol and message templates, see [references/details.md](references/details.md).

## What Workers Do (Controller Reference)

Workers follow a TDD cycle upon receiving a `[New Change Assigned]` notification: read requirements → write acceptance tests → implement (marking tasks done) → run verify → notify main with result.

Workers notify the controller using `reply-main` with one of:
- `"Task completed: <summary>; verify: passed"`
- `"Task completed: <summary>; verify: failed — <reason>"`
- `"Task completed: <summary>; verify: skipped"`
- `"Need decision: <problem or options>"` (when blocked)

Full worker workflow details: [references/worker-workflow.md](references/worker-workflow.md)

## Brainstorming (Required Before Assign)

<HARD-GATE>
Do NOT execute `agent-team worker assign`, write any code, or take any implementation action
until you have presented a design and the user has explicitly approved it.
This applies to EVERY assignment regardless of perceived simplicity.
</HARD-GATE>

When the user intends to assign new work to a worker, you MUST follow the brainstorming process.

For the full checklist, principles, and anti-patterns, see [references/brainstorming.md](references/brainstorming.md).

## Backend Selection

The default terminal backend is **WezTerm**. To use **tmux** instead, set the environment variable:

```bash
AGENT_TEAM_BACKEND=tmux agent-team <command>
```
