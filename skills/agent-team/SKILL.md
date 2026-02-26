---
name: agent-team
description: >
  AI team role and worker manager for multi-agent development workflows.
  Uses Role (skill package) + Worker (instance) model with git worktrees.
  Use when the user wants to create roles, manage workers, assign tasks,
  check team status, or merge worker branches.
  Triggers on /agent-team commands, "create a role", "create a worker",
  "open worker session", "assign task", "show team status", "merge worker branch".
---

# agent-team

Manages AI team workflows using a **Role + Worker** dual-layer model with git worktrees and terminal multiplexer tabs.

- **Role** = a skill package definition (SKILL.md + system.md + role.yaml) stored in `agents/teams/<role-name>/`
- **Worker** = a role instance running in an isolated worktree, identified by `<role-name>-<3-digit-number>`

For directory layout details, see [references/details.md](references/details.md).

## Install

```bash
brew tap JsonLee12138/agent-team && brew install agent-team
```

## Upgrade

```bash
brew update && brew upgrade agent-team
```

## Role Management (AI Workflow)

Roles are created and managed by AI using the **role-creator** skill. The CLI does not handle role creation.

### Creating a Role

1. Use the `/role-creator` skill with `--target-dir agents/teams`
2. If a role already exists in global `~/.claude/skills/`, prompt user to copy it to `agents/teams/`
3. Result: `agents/teams/<role-name>/` with SKILL.md, system.md, references/role.yaml

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

1. Verifies role exists in `agents/teams/<role-name>/`
2. If not found, checks global skills and offers to copy
3. Creates worktree `.worktrees/<worker-id>/` with branch `team/<worker-id>`
4. Generates `.gitignore` (excludes `.gitignore`, `.claude/`, `.codex/`, `openspec/`)
5. Creates `agents/workers/<worker-id>/config.yaml`
6. Initializes OpenSpec

### Open a worker session

```bash
agent-team worker open <worker-id> [claude|codex|opencode] [--model <model>] [--new-window]
```

1. Reads role from worker config
2. Copies role skill and dependency skills to `.claude/skills/` and `.codex/skills/` (mirrored)
3. Generates CLAUDE.md from role's system.md
4. Opens terminal tab with chosen AI provider
5. `--new-window` / `-w`: Open in a new WezTerm window

### Assign a change

```bash
agent-team worker assign <worker-id> "<description>" [provider] [--proposal <file>] [--design <file>] [--new-window]
```

1. Creates an OpenSpec change at `openspec/changes/<timestamp>-<slug>/`
2. Copies `--design` file as `design.md` (brainstorming output)
3. Copies `--proposal` file as `proposal.md` (work requirements)
4. Auto-opens the worker session if not running
5. Sends a `[New Change Assigned]` notification

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

Sends a message prefixed with `[Worker: <worker-id>]` to the controller's session.

## Brainstorming (Required Before Assign)

<HARD-GATE>
Do NOT execute `agent-team worker assign`, write any code, or take any implementation action
until you have presented a design and the user has explicitly approved it.
This applies to EVERY assignment regardless of perceived simplicity.
</HARD-GATE>

When the user intends to assign new work to a worker, you MUST follow the brainstorming process.

For the full checklist, principles, and anti-patterns, see [references/brainstorming.md](references/brainstorming.md).

## Task Completion Rules

Workers MUST follow these rules when completing a task:

1. **Archive the change**: Run `/openspec archive` to mark the change as completed
2. **Notify the controller**: Run `agent-team reply-main "<summary>"` to notify the main controller (unless explicitly told not to)
3. **New work requires new tasks**: After archiving, any new work must go through a new assign cycle

## Backend Selection

Use tmux instead of WezTerm (default):

```bash
AGENT_TEAM_BACKEND=tmux agent-team <command>
```
