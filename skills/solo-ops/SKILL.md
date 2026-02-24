---
name: solo-ops
description: >
  AI team role manager for multi-agent development workflows.
  Use when the user wants to create/delete team roles, open role sessions in WezTerm tabs,
  assign tasks to roles, check team status, or merge role branches.
  Triggers on /solo-ops commands, "create a team role", "open role session",
  "assign task to role", "show team status", "merge role branch".
---

# solo-ops

Manages AI team roles using git worktrees + terminal multiplexer tabs. Each role runs in its own isolated worktree (branch `team/<name>`) and opens as a full-permission AI session in a new tab.

For directory layout, task format, and bidirectional communication details, see [references/details.md](references/details.md).

## Install

```bash
brew tap leeforge/tap && brew install agent-team
```

## Usage

Run from within the project git repository:

```bash
agent-team <command>
```

Use tmux backend by setting an environment variable:

```bash
AGENT_TEAM_BACKEND=tmux agent-team <command>
```

## Commands

Always run from within the project git repository.

### Create a role
```bash
agent-team create <name>
```
Creates `team/<name>` git branch + worktree at `.worktrees/<name>/`. Generates:
- `agents/teams/<name>/config.yaml` — provider, description, pane tracking
- `agents/teams/<name>/prompt.md` — role system prompt (edit this to define the role)
- `agents/teams/<name>/tasks/pending/` and `tasks/done/`

After creating, guide the user to edit `prompt.md` to define the role's expertise and behavior.

### Open a role session
```bash
agent-team open <name> [claude|codex|opencode] [--model <model>]
```
- Copies `prompt.md` → `CLAUDE.md` in worktree root (auto-injected as system context)
- Spawns a new terminal tab titled `<name>` running `claude --dangerously-skip-permissions` (or `codex --dangerously-bypass-approvals-and-sandbox`)
- Provider priority: argument > `config.yaml default_provider` > claude

### Open all sessions
```bash
agent-team open-all [claude|codex|opencode] [--model <model>]
```
Opens every role that has a config.yaml.

### Assign a task
```bash
agent-team assign <name> "<task description>" [claude|codex|opencode] [--model <model>]
```
1. Writes `agents/teams/<name>/tasks/pending/<timestamp>-<slug>.md`
2. Auto-opens the role session if not running
3. Sends a notification message to the running session

### Reply to a role
```bash
agent-team reply <name> "<answer>"
```
Sends a reply to a role's running session. Used when a role has asked a question via `ask claude` and the main controller wants to respond. The message is prefixed with `[Main Controller Reply]` so the role AI can identify it.

### Check status
```bash
agent-team status
```
Shows all roles, whether their session is running (by pane-id), and pending task count.

### Merge completed work
```bash
agent-team merge <name>
```
Merges `team/<name>` into the current branch with `--no-ff`. Run `delete` afterward to clean up.

### Delete a role
```bash
agent-team delete <name>
```
Removes the worktree and deletes the `team/<name>` branch.
