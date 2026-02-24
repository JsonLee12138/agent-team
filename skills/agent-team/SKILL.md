---
name: agent-team
description: >
  AI team role manager for multi-agent development workflows.
  Use when the user wants to create/delete team roles, open role sessions in terminal tabs,
  assign tasks to roles, check team status, or merge role branches.
  Triggers on /agent-team commands, "create a team role", "open role session",
  "assign task to role", "show team status", "merge role branch".
---

# agent-team

Manages AI team roles using git worktrees + terminal multiplexer tabs. Each role runs in its own isolated worktree (branch `team/<name>`) and opens as a full-permission AI session in a new tab.

For directory layout and bidirectional communication details, see [references/details.md](references/details.md).

## Install

```bash
brew tap JsonLee12138/agent-team && brew install agent-team
```

## Usage

Run from within a project git repository:

```bash
agent-team <command>
```

Use tmux backend (default is WezTerm):

```bash
AGENT_TEAM_BACKEND=tmux agent-team <command>
```

## Brainstorming (Required Before Assign)

When the user intends to assign new work to a role, you MUST brainstorm first:

1. **Explore context** — check the role's `prompt.md`, existing `openspec/specs/`, and project state
2. **Ask clarifying questions** — one at a time, prefer multiple choice when possible
3. **Propose 2-3 approaches** — with trade-offs and your recommendation
4. **User confirms design** — get explicit approval
5. **Write proposal** — save the confirmed design to a temp file
6. **Execute assign** — run `agent-team assign <name> "<desc>" --proposal <file>`

**Rules:**
- Brainstorming is **mandatory** for new work assignments
- Can be skipped when user explicitly says "just assign" or provides a complete design
- One question at a time. YAGNI. Explore alternatives before settling.

## Commands

### Create a role
```bash
agent-team create <name>
```
Creates `team/<name>` git branch + worktree at `.worktrees/<name>/`. Generates:
- `agents/teams/<name>/config.yaml` — provider, description, pane tracking
- `agents/teams/<name>/prompt.md` — role system prompt (edit this to define the role)
- `openspec/` — OpenSpec project structure for change management

After creating, guide the user to edit `prompt.md` to define the role's expertise and behavior.

### Open a role session
```bash
agent-team open <name> [claude|codex|opencode] [--model <model>]
```
- Generates `CLAUDE.md` in worktree root from `prompt.md` (auto-injected as system context)
- Spawns a new terminal tab titled `<name>` running the chosen AI provider
- Provider priority: CLI argument > `config.yaml default_provider` > claude
- Model priority: `--model` flag > `config.yaml default_model` > provider default

### Open all sessions
```bash
agent-team open-all [claude|codex|opencode] [--model <model>]
```
Opens every role that has a config.yaml.

### Assign a change
```bash
agent-team assign <name> "<description>" [claude|codex|opencode] [--model <model>] [--proposal <file>]
```
1. Creates an OpenSpec change at `openspec/changes/<timestamp>-<slug>/`
2. Writes the proposal file from `--proposal` flag (or empty if not provided)
3. Auto-opens the role session if not running
4. Sends a `[New Change Assigned]` notification to the running session

The role will then use `/opsx:continue` to proceed through specs → design → tasks → apply.

### Reply to a role
```bash
agent-team reply <name> "<answer>"
```
Sends a reply to a role's running session, prefixed with `[Main Controller Reply]`.

### Check status
```bash
agent-team status
```
Shows all roles, session status (running/stopped), and active OpenSpec changes.

### Merge completed work
```bash
agent-team merge <name>
```
Merges `team/<name>` into the current branch with `--no-ff`. Run `delete` afterward to clean up.

### Delete a role
```bash
agent-team delete <name>
```
Closes the running session (if any), removes the worktree, and deletes the `team/<name>` branch.
