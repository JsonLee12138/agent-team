# agent-team

English | [中文](./README.zh.md)

AI team role and worker manager for multi-agent development workflows. It uses a **Role + Worker** model with isolated Git worktrees and terminal sessions.

- **Role**: skill package definition in `agents/teams/<role-name>/`
- **Worker**: runtime role instance in `.worktrees/<worker-id>/` (for example `frontend-dev-001`)

## Table of Contents

- [How It Works](#how-it-works)
- [Requirements](#requirements)
- [Installation](#installation)
- [Upgrade](#upgrade)
- [Quick Start](#quick-start)
- [Built-in Roles](#built-in-roles)
- [Usage as a Skill](#usage-as-a-skill)
- [CLI Reference](#cli-reference)
- [Directory Structure](#directory-structure)
- [Supported Providers](#supported-providers)
- [Environment Variables](#environment-variables)
- [License](#license)

## How It Works

```
Main branch
  ├── agents/teams/<role-name>/          <- role skill definitions
  ├── agents/workers/<worker-id>/        <- worker config
  └── .worktrees/<worker-id>/            <- isolated runtime workspace
```

Typical flow:

1. Create or prepare a role in `agents/teams/`
2. Create worker: `agent-team worker create <role-name>`
3. Open worker session: `agent-team worker open <worker-id> [provider]`
4. Brainstorm, then assign change: `agent-team worker assign ...`
5. Merge: `agent-team worker merge <worker-id>`
6. Cleanup: `agent-team worker delete <worker-id>`

## Requirements

- Git
- [WezTerm](https://wezfurlong.org/wezterm/) or [tmux](https://github.com/tmux/tmux)
- At least one AI provider CLI: [claude](https://github.com/anthropics/claude-code), [codex](https://github.com/openai/codex), or [opencode](https://opencode.ai)
- [Node.js](https://nodejs.org/) (for OpenSpec auto-install during worker creation)

## Installation

### Agent Skill (recommended)

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

Download a binary from [Releases](https://github.com/JsonLee12138/agent-team/releases), extract it, and add it to your `PATH`.

## Upgrade

```bash
# Skill
npx skills add JsonLee12138/agent-team

# Homebrew
brew update && brew upgrade agent-team

# Source
go install github.com/JsonLee12138/agent-team@latest
```

## Quick Start

1. Create role(s) with `role-creator` into `agents/teams/`
```bash
python3 skills/role-creator/scripts/create_role_skill.py \
  --repo-root . \
  --role-name frontend-dev \
  --target-dir agents/teams \
  --description "Frontend role for UI implementation" \
  --system-goal "Ship maintainable frontend features"
```

2. List roles
```bash
agent-team role list
```

3. Create worker
```bash
agent-team worker create frontend-dev
```

4. Open worker session
```bash
agent-team worker open frontend-dev-001 claude
```

5. Assign change
```bash
agent-team worker assign frontend-dev-001 "Implement responsive navbar"
```

6. Merge and delete worker
```bash
agent-team worker merge frontend-dev-001
agent-team worker delete frontend-dev-001
```

## Built-in Roles

This repository currently includes one built-in role:

- `frontend-architect` (path: `skills/frontend-architect/`)

To use it with `agent-team`, copy it into `agents/teams/` first:

```bash
mkdir -p agents/teams
cp -R skills/frontend-architect agents/teams/
agent-team role list
```

## Usage as a Skill

With agent skill installed, you can describe intent in natural language:

- "Create a team role for frontend architecture."
- "Create a worker for frontend-architect and open it with codex."
- "Assign a change to frontend-architect-001."
- "Show worker status."

The controller should run brainstorming before assignment, then create an OpenSpec change and notify the worker session.

## CLI Reference

All commands run inside a Git repository.

### Role commands

| Command | Description |
|---------|-------------|
| `agent-team role list` | List available roles in `agents/teams/` |

### Worker commands

| Command | Description |
|---------|-------------|
| `agent-team worker create <role-name>` | Create worker worktree, branch, config, and initialize OpenSpec |
| `agent-team worker open <worker-id> [provider] [--model <model>] [--new-window]` | Open worker session |
| `agent-team worker assign <worker-id> "<description>" [provider] [--proposal <file>] [--design <file>] [--model <model>] [--new-window]` | Create OpenSpec change and notify worker |
| `agent-team worker status` | Show workers, roles, running state, skills, and active changes |
| `agent-team worker merge <worker-id>` | Merge `team/<worker-id>` into current branch |
| `agent-team worker delete <worker-id>` | Delete worker worktree, branch, and config |

### Communication commands

| Command | Description |
|---------|-------------|
| `agent-team reply <worker-id> "<answer>"` | Send `[Main Controller Reply]` message to a worker session |
| `agent-team reply-main "<message>"` | Worker sends `[Worker: <worker-id>]` message to controller session |

Use `AGENT_TEAM_BACKEND=tmux` before commands to switch backend from WezTerm to tmux.

## Directory Structure

```
project-root/
├── agents/
│   ├── teams/
│   │   └── <role-name>/
│   │       ├── SKILL.md
│   │       ├── system.md
│   │       └── references/role.yaml
│   └── workers/
│       └── <worker-id>/config.yaml
└── .worktrees/
    └── <worker-id>/
        ├── .claude/skills/
        ├── .codex/skills/
        ├── CLAUDE.md
        ├── AGENTS.md
        └── openspec/
            ├── specs/
            └── changes/
```

## Supported Providers

| Provider | Value |
|----------|-------|
| Claude Code | `claude` (default) |
| OpenAI Codex | `codex` |
| OpenCode | `opencode` |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `AGENT_TEAM_BACKEND` | Terminal backend: `wezterm` (default) or `tmux` |

## License

MIT
