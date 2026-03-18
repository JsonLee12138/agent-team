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

# agent-team

Manages AI team workflows using a **Role + Worker** dual-layer model with git worktrees and terminal multiplexer tabs.

- **Role** = a skill package definition (SKILL.md + system.md + role.yaml) stored in `.agents/teams/<role-name>/` (project) or `~/.agents/roles/<role-name>/` (global)
- **Worker** = a role instance running in an isolated worktree, identified by `<role-name>-<3-digit-number>`

For directory layout details, see [references/details.md](references/details.md).

> **IMPORTANT -- No Sandbox Mode**
> All `agent-team` commands MUST be run **without sandbox** (i.e. with `dangerouslyDisableSandbox: true` or equivalent).
> agent-team manages git worktrees, creates/deletes directories outside the project root, spawns terminal sessions, and writes to `~/.agents/`. These operations will fail silently or error out inside a sandboxed shell.

## Install & Upgrade

See [references/install-upgrade.md](references/install-upgrade.md). If `agent-team` binary is not found, follow the install guide.

> **Strategy**: Plugin platforms (Claude Code, Gemini CLI, OpenCode) install/update via their native plugin system only — do NOT use `npx skills add`. Codex uses `npx skills add -a codex -g -y` (global only). Other non-plugin platforms use `npx skills add -a <platform>`.

## Phase 1: Initialization

```bash
agent-team init [--global-only] [--skip-detect]
```

Runs first-time setup: detects installed providers, creates `.agents/teams/` structure, and installs bundled roles to `~/.agents/roles/`.

> **AI Behavior**: Detect missing `.agents/teams/` directory and prompt user to run `agent-team init`. Do NOT auto-run init.

## Phase 2: Role Preparation

### Creating a Role

> **MANDATORY**: Role creation MUST be delegated to the `/role-creator` skill via `Skill` tool invocation. Do NOT create role files (SKILL.md, system.md, role.yaml) manually or inline. Always invoke `/role-creator` first, then continue the agent-team workflow.

1. **Invoke `/role-creator`** skill with `--target-dir .agents/teams` (use the `Skill` tool with `skill: "role-creator"`)
2. If a role already exists in global `~/.agents/roles/`, prompt user to copy it to `.agents/teams/`
3. Result: `.agents/teams/<role-name>/` with SKILL.md, system.md, references/role.yaml

### Creating a Team (Batch Role Creation)

When the user describes a team in natural language (e.g. "Create a frontend developer, a QA engineer, and a frontend architect role"):

1. Parse the prompt into responsibility units (one role per responsibility).
2. Normalize each role name to kebab-case (`frontend developer` -> `frontend-dev`).
3. Present the draft role list and ask for one confirmation before execution.
4. For each approved role, run the full **Creating a Role** workflow — MUST invoke `/role-creator` skill via `Skill` tool (do NOT create role files manually).
5. Return a per-role summary: `created`, `already exists`, or `failed` (with reason).
6. Stop after role creation. Do NOT create workers in this flow.

Rules:
- Team creation is role-only. MUST NOT auto-run `worker create`, `worker open`, or `worker assign`.
- If a role already exists, do not overwrite. Mark as `already exists`.
- A single role failure does not cancel the batch. Continue and report final results.

### Role Discovery Flow

When a role is not found locally:

1. Ask user: "Create from scratch?" or "Search Role Hub?"
2. If search: `agent-team catalog search <query>` -> show results -> `agent-team role-repo add <source> --role <name>`
3. If no match or declined: fall back to **Creating a Role** (which MUST invoke `/role-creator` skill)

### Role Repo Management

```bash
# Search GitHub for role repositories
agent-team role-repo find <query>

# Install roles from a repository
agent-team role-repo add <owner/repo> [--role <name>] [-g] [-y]

# List installed repository-managed roles
agent-team role-repo list [-g] [--json]

# Check installed roles for remote updates
agent-team role-repo check

# Update installed roles from remote sources
agent-team role-repo update

# Remove installed repository-managed roles
agent-team role-repo remove <name>
```

### Listing Local Roles

```bash
agent-team role list
```

## Phase 3: Worker Lifecycle

### Create a worker

```bash
agent-team worker create <role-name> [--provider <provider>] [--model <model>]
```

Creates worktree `.worktrees/<worker-id>/` with branch `team/<worker-id>`, writes `worker.yaml`, initializes `.tasks/` directory, syncs skills, and injects role prompt. Does **not** open a terminal session — use `worker open` to start the session. If `--provider` is omitted, `worker.yaml.provider` defaults to `claude`.

> **AI Behavior (Skill Sync Warning)**: If skill installation emits a warning during `worker create`, surface the warning to the user and ask whether role skill bindings need adjustment. Do NOT auto-fix bindings.

### Open a worker session

```bash
agent-team worker open <worker-id> [--provider <provider>] [--model <model>] [--new-window]
```

Copies role + dependency skills into the worktree, generates CLAUDE.md/AGENTS.md from role's system.md, and opens a terminal tab with the chosen AI provider. `worker.yaml` remains the persisted source of `provider` and `default_model`: `--provider` / `--model` update config when passed; omitted flags reuse existing config. `--new-window` is optional. See [references/platforms.md](references/platforms.md).

> **AI Behavior (Session Start)**: Ensure worker loads role context and is aware of pending tasks.

### Assign a change

```bash
agent-team worker assign <worker-id> "<description>" [provider] [--proposal <file>] [--design <file>] [--verify-cmd <cmd>] [--new-window]
```

Creates a task change at `.tasks/changes/<timestamp>-<slug>/` and sends a `[New Change Assigned]` notification. The CLI auto-opens the worker session if not running.

**Before running this command**, the controller MUST:

1. Complete the brainstorming process -> [references/brainstorming.md](references/brainstorming.md)
2. Pass the Assign Readiness Gate -> [references/readiness.md](references/readiness.md)
3. **Sync worktree with main**: If worker exists with no uncommitted changes, rebase/merge latest main. If uncommitted changes exist, ask user first.

After a successful assign, the controller's default next state is `waiting for worker reply`.

- Do NOT enter a routine `worker status` polling loop.
- Wait for the worker to report back with `agent-team reply-main "<message>"`.
- Use `worker status` only for assign failure diagnosis, timeout or no-reply investigation, or when the user explicitly asks for current status.

### Check status

```bash
agent-team worker status
```

Shows all workers, their roles, session status, and active changes.

> **AI Behavior (Status Usage)**: `worker status` is an exception and inspection tool. It is not the default next step after `worker assign`.

### Merge completed work

```bash
agent-team worker merge <worker-id>
```

Merges `team/<worker-id>` into the current branch with `--no-ff`. Does **not** close the worker session — use `worker close` explicitly if needed before or after merge.

### Close a session

```bash
agent-team worker close <worker-id>
```

Closes the terminal pane without deleting the worktree or branch. Idempotent — succeeds if the session is already closed. Can reopen later with `worker open`.

### Delete a worker

```bash
agent-team worker delete <worker-id>
```

Closes the running session, removes the worktree, deletes the branch, and cleans up config. Stops if the session close fails (will not delete a worktree with a live session). **Irreversible.**

### AI Behavior -- Worker Cleanup Rules

- After task completion notification: do NOT auto-merge. Ask user: "Review code, merge, or assign next task?"
- After merge: do NOT auto-delete. Ask user: "Keep worker or delete?"
- NEVER run `worker delete` without explicit user approval
- Closing a session does NOT require confirmation

### AI Behavior -- Idle Worker Scheduling

- Inform user when a worker is idle with pending tasks. Do NOT auto-assign.

## Phase 4: Task & Monitoring

### Task Commands

```bash
agent-team task create <worker-id> "<description>" [--proposal <file>] [--design <file>] [--verify-cmd <cmd>] [--skip-verify]
agent-team task list [--worker <worker-id>] [--status <status>]
agent-team task show <worker-id> <change-name>
agent-team task done <worker-id> <change-name> <task-id>
agent-team task verify <worker-id> <change-name>
agent-team task archive <worker-id> <change-name>
```

### Worker TDD Cycle

Workers follow a TDD cycle upon receiving a `[New Change Assigned]` notification: read requirements -> write acceptance tests -> implement (marking tasks done) -> run verify -> notify main with result.

Full worker workflow details: [references/worker-workflow.md](references/worker-workflow.md)

### Worker Notification Formats

Workers notify the controller using `reply-main` with one of:
- `"Task completed: <summary>; change archived: <change-name>"`
- `"Task completed: <summary>; archive failed for <change-name>: <error>"`
- `"Need decision: <problem or options>"` (when blocked)

> **AI Behavior (Task Completion)**: Treat worker `reply-main` as the default completion signal after assignment, surface the outcome to the user, and ask how to proceed. Do NOT auto-merge.

## Phase 5: Communication

```bash
# Controller -> Worker
agent-team reply <worker-id> "<answer>"

# Worker -> Controller
agent-team reply-main "<message>"
```

Messages appear in the controller's terminal as `[Worker: <worker-id>] <message>`. For full protocol details, see [references/details.md](references/details.md).

> **AI Behavior**: Surface worker questions to user immediately. Do NOT answer on behalf of user. Batch-present multiple blocked workers.

## Skill Cache Management

Remote skills installed via `npx skills add` are cached at `.agents/.cache/skills/` and symlinked into worktrees. Use the `skill` subcommands to manage this cache.

```bash
agent-team skill check                # Check installed skills for available updates
agent-team skill update               # Update all cached skills to latest versions
agent-team skill clean [--force]      # Remove cached skills
```

- `skill clean` checks whether cached skills are actively symlinked by existing worktrees before removing them.
  - Unused skills are removed immediately.
  - In-use skills are listed with their worker IDs and require user confirmation before removal.
  - Use `--force` to skip the confirmation prompt and remove all cached skills (including those in use).
- `worker create --fresh` forces re-installation of all skills, bypassing the project cache.

## Catalog (Role Hub)

Query commands for browsing the role catalog:

```bash
agent-team catalog search <query>     # Search verified roles
agent-team catalog show <name>        # Show role details
agent-team catalog list               # List all roles
agent-team catalog repo <owner/repo>  # Show roles from a repository
agent-team catalog stats              # Show catalog statistics
```

Admin commands (`discover`, `normalize`, `serve`) are available via `agent-team catalog --help`.

Links to **Phase 2: Role Discovery Flow** for installing found roles.

## Backend Selection

Default terminal backend is **WezTerm**. To use **tmux**:

```bash
AGENT_TEAM_BACKEND=tmux agent-team <command>
```

## Migration

```bash
agent-team migrate
```

Migrates legacy `agents/` directory to `.agents/`.
