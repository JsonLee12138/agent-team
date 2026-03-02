# agent-team

English | [中文](./README.zh.md)

AI team role and worker manager for multi-agent development workflows. It uses a **Role + Worker** model with isolated Git worktrees and terminal sessions.

- **Role**: skill package definition in `.agents/teams/<role-name>/`
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
- [Role Repo Locks](#role-repo-locks)
- [Directory Structure](#directory-structure)
- [Supported Providers](#supported-providers)
- [Environment Variables](#environment-variables)
- [License](#license)

## How It Works

```
Main branch
  ├── .agents/teams/<role-name>/           <- role skill definitions
  └── .worktrees/<worker-id>/              <- isolated runtime workspace
        └── worker.yaml                    <- worker config
```

Typical flow:

1. Create or prepare a role in `.agents/teams/`
2. Create and open worker: `agent-team worker create <role-name> [provider]`
3. Brainstorm, then assign change: `agent-team worker assign ...`
4. Merge: `agent-team worker merge <worker-id>`
5. Cleanup: `agent-team worker delete <worker-id>`

## Requirements

- Git
- [WezTerm](https://wezfurlong.org/wezterm/) or [tmux](https://github.com/tmux/tmux)
- At least one AI provider CLI: [claude](https://github.com/anthropics/claude-code), [codex](https://github.com/openai/codex), or [opencode](https://opencode.ai)
- [Node.js](https://nodejs.org/) (for OpenSpec auto-install and `npx skills add` during worker creation)

## Installation

**Note:** Installation differs by platform. Claude Code has a built-in plugin system. Other platforms use the Agent Skill method.

### Claude Code (via Plugin Marketplace)

```bash
# 1. Add marketplace
/plugin marketplace add JsonLee12138/agent-team

# 2. Install plugin
/plugin install agent-team@agent-team
```

Or via CLI:

```bash
claude plugin marketplace add JsonLee12138/agent-team
claude plugin install agent-team@agent-team
```

### Agent Skill

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
# Plugin
/plugin marketplace update agent-team
# or
claude plugin marketplace update agent-team

# Skill
npx skills add JsonLee12138/agent-team

# Homebrew
brew update && brew upgrade agent-team

# Source
go install github.com/JsonLee12138/agent-team@latest
```

## Quick Start

1. Create role(s) with `role-creator` into `.agents/teams/`
```bash
python3 skills/role-creator/scripts/create_role_skill.py \
  --repo-root . \
  --role-name frontend-dev \
  --target-dir .agents/teams \
  --description "Frontend role for UI implementation" \
  --system-goal "Ship maintainable frontend features"
```

2. List roles
```bash
agent-team role list
```

3. Create and open worker (creates worktree, opens session, installs skills, launches AI)
```bash
agent-team worker create frontend-dev claude
```

4. Assign change
```bash
agent-team worker assign frontend-dev-001 "Implement responsive navbar"
```

5. Merge and delete worker
```bash
agent-team worker merge frontend-dev-001
agent-team worker delete frontend-dev-001
```

## Built-in Roles

This repository currently includes one built-in role:

- `frontend-architect` (path: `skills/frontend-architect/`)

To use it with `agent-team`, copy it into `.agents/teams/` first:

```bash
mkdir -p .agents/teams
cp -R skills/frontend-architect .agents/teams/
agent-team role list
```

## Built-in Skills

| Skill | Description |
|-------|-------------|
| `role-creator` | Create or update role skill packages interactively |
| `brainstorming` | Turn rough ideas into validated design docs through one-question-at-a-time dialogue before implementation |

## Usage as a Skill

With agent skill installed, you can describe intent in natural language:

- "Create a team role for frontend architecture."
- "Create a worker for frontend-architect with codex."
- "Assign a change to frontend-architect-001."
- "Show worker status."

The controller should run brainstorming before assignment, then create an OpenSpec change and notify the worker session.

## CLI Reference

All commands run inside a Git repository.

### Role commands

| Command | Description |
|---------|-------------|
| `agent-team role list` | List available roles in `.agents/teams/` |

### Role Repository commands

| Command | Description |
|---------|-------------|
| `agent-team role-repo search <query>` | Search GitHub roles using strict role path contracts |
| `agent-team role-repo add <source> [--role <name>...] [--list] [-g] [-y]` | Discover and install role(s) from `owner/repo` or GitHub URL (interactive selector when multiple roles are found) |
| `agent-team role-repo list [-g]` | List installed repository-managed roles in selected scope |
| `agent-team role-repo remove [roles...] [-g] [-y]` | Remove installed roles and clean lock entries (interactive selector/confirm by default) |
| `agent-team role-repo check [-g]` | Check lock entries against remote folder hashes |
| `agent-team role-repo update [-g] [-y]` | Update roles with remote changes (interactive selector by default, `-y` updates all) |

Accepted remote role path contracts:

- `skills/<role>/references/role.yaml`
- `.agents/teams/<role>/references/role.yaml`

### Worker commands

| Command | Description |
|---------|-------------|
| `agent-team worker create <role-name> [provider] [--model <model>] [--new-window]` | Create worker, open session, install skills, and launch AI |
| `agent-team worker open <worker-id> [provider] [--model <model>] [--new-window]` | Reopen an existing worker session |
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

## Role Repo Locks

- Project lock: `roles-lock.json`
- Global lock: `~/.agents/.role-lock.json`
- Project install target: `.agents/teams/<role>/`
- Global install target: `~/.agents/roles/<role>/`

## Directory Structure

```
project-root/
├── .agents/
│   └── teams/
│       └── <role-name>/
│           ├── SKILL.md
│           ├── system.md
│           └── references/role.yaml
└── .worktrees/
    └── <worker-id>/
        ├── worker.yaml              <- worker config
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
