# agent-team

English | [中文](./README.zh.md)

AI team role manager for multi-agent development workflows. Orchestrate multiple AI coding agents in isolated Git worktrees with WezTerm or tmux.

Each role gets its own Git branch, worktree, terminal session, and task inbox — all managed through natural language prompts or CLI commands.

## Table of Contents

- [How It Works](#how-it-works)
- [Requirements](#requirements)
- [Installation](#installation)
  - [Agent Skill (recommended)](#agent-skill-recommended)
  - [Homebrew (macOS)](#homebrew-macos)
  - [From Source](#from-source)
  - [From GitHub Releases](#from-github-releases)
- [Upgrade](#upgrade)
- [Usage as a Skill](#usage-as-a-skill)
  - [Create a role](#create-a-role)
  - [Open a role session](#open-a-role-session)
  - [Assign a task](#assign-a-task)
  - [Reply to a role](#reply-to-a-role)
  - [Check status](#check-status)
  - [Merge and clean up](#merge-and-clean-up)
- [CLI Reference](#cli-reference)
- [Directory Structure](#directory-structure)
- [Supported Providers](#supported-providers)
- [Environment Variables](#environment-variables)
- [License](#license)

## How It Works

```
Main branch
    │
    ├── .worktrees/frontend/     ← team/frontend branch + Claude session
    ├── .worktrees/backend/      ← team/backend branch + Codex session
    └── .worktrees/qa/           ← team/qa branch + OpenCode session
```

Each role runs in isolation. You assign tasks through the main agent, roles work independently, you merge when done.

## Requirements

- Git
- [WezTerm](https://wezfurlong.org/wezterm/) or [tmux](https://github.com/tmux/tmux)
- At least one AI provider CLI: [claude](https://github.com/anthropics/claude-code), [codex](https://github.com/openai/codex), or [opencode](https://opencode.ai)

## Installation

### Agent Skill (recommended)

Install as an Agent Skill so Claude Code (or any compatible AI agent) can manage your team through natural language:

```bash
npx skills add JsonLee12138/agent-team
```

### Homebrew (macOS)

```bash
brew tap JsonLee12138/agent-team
brew install agent-team
```

### From Source

Requires Go 1.24+.

```bash
go install github.com/JsonLee12138/agent-team@latest
```

### From GitHub Releases

Download the binary for your platform from [Releases](https://github.com/JsonLee12138/agent-team/releases), extract it, and add it to your `PATH`.

## Upgrade

### Agent Skill

```bash
npx skills add JsonLee12138/agent-team
```

### Homebrew

```bash
brew update && brew upgrade agent-team
```

### From Source

```bash
go install github.com/JsonLee12138/agent-team@latest
```

## Usage as a Skill

Once installed, just describe what you want in natural language inside your AI agent session. No need to remember command syntax.

### Create a role

> Create a team role called "frontend" for building React components.

> Create a backend role that handles API development with Node.js.

The agent will scaffold the worktree and prompt you to define the role's expertise in `prompt.md`.

---

### Open a role session

> Open the frontend role with Claude.

> Open all role sessions using Codex.

> Open the backend role with model claude-opus-4-5.

A new terminal tab (or tmux window) opens with the AI provider running inside the role's worktree.

---

### Assign a task

> Assign a task to frontend: implement a responsive navbar with a mobile hamburger menu.

> Tell the backend role to add a JWT authentication middleware.

The task is written to the role's pending inbox and the session is notified. If the session isn't running, it starts automatically.

---

### Reply to a role

> Reply to frontend: use CSS Grid for the layout, Flexbox for individual items.

> Tell the backend role that we're using PostgreSQL, not MySQL.

The message is delivered to the running session prefixed with `[Main Controller Reply]`.

---

### Check status

> Show team status.

> Which roles are currently running?

Displays all roles, their session state, and pending task count.

---

### Merge and clean up

> Merge the frontend branch.

> Delete the backend role after merging.

Merges `team/<name>` into the current branch with `--no-ff`, then optionally removes the worktree and branch.

---

## CLI Reference

All commands run from within a Git repository.

| Command | Description |
|---------|-------------|
| `agent-team create <name>` | Create a new role (branch + worktree + scaffold) |
| `agent-team open <name> [provider] [--model <m>]` | Open a role session in a new terminal tab |
| `agent-team open-all [provider] [--model <m>]` | Open sessions for all roles |
| `agent-team assign <name> "<task>" [provider]` | Assign a task and notify the session |
| `agent-team reply <name> "<message>"` | Send a reply to a running role session |
| `agent-team status` | Show all roles, status, and pending tasks |
| `agent-team merge <name>` | Merge role branch into current branch |
| `agent-team delete <name>` | Close session, remove worktree and branch |

Use `AGENT_TEAM_BACKEND=tmux` before any command to switch to the tmux backend.

## Directory Structure

```
project-root/
└── .worktrees/
    └── <name>/
        ├── CLAUDE.md                              ← auto-generated from prompt.md
        └── agents/teams/<name>/
            ├── config.yaml                        ← provider, model, pane_id
            ├── prompt.md                          ← role identity (edit this)
            └── tasks/
                ├── pending/<timestamp>-<slug>.md  ← active tasks
                └── done/<timestamp>-<slug>.md     ← completed tasks
```

## Supported Providers

| Provider | Value |
|----------|-------|
| Claude Code | `claude` (default) |
| OpenAI Codex | `codex` |
| OpenCode | `opencode` |

Override per-role in `config.yaml` (`default_provider`) or pass as a positional argument to `open` / `assign`.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `AGENT_TEAM_BACKEND` | Terminal backend: `wezterm` (default) or `tmux` |

## License

MIT
