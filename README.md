# agent-team

English | [中文](./README.zh.md)

AI team role and worker manager for multi-agent development workflows. It uses a **Role + Worker** model with isolated Git worktrees and terminal sessions.

- **Role**: skill package definition in `.agents/teams/<role-name>/` (project) or `~/.agents/roles/<role-name>/` (global)
- **Worker**: runtime role instance in `.worktrees/<worker-id>/` (for example `frontend-dev-001`)

## Table of Contents

- [How It Works](#how-it-works)
- [Requirements](#requirements)
- [Installation](#installation)
- [Upgrade](#upgrade)
- [Quick Start](#quick-start)
- [Built-in Roles](#built-in-roles)
- [Built-in Skills](#built-in-skills)
- [Supported Providers](#supported-providers)
- [Advanced](#advanced)
- [License](#license)

## How It Works

```
Main branch
  ├── .agents/teams/<role-name>/           <- project role definitions
  └── .worktrees/<worker-id>/              <- isolated runtime workspace
        └── worker.yaml                    <- worker config

~/.agents/roles/<role-name>/               <- global role definitions
```

Typical flow:

1. **Define roles** — "Create a frontend developer role for my project."
2. **Spawn worker** — "Create a worker for frontend-dev with claude."
3. **Brainstorm & assign** — "Assign frontend-dev-001 to implement responsive navbar."
4. **Merge results** — "Merge frontend-dev-001."
5. **Cleanup** — "Delete worker frontend-dev-001."

## Requirements

- Git
- [WezTerm](https://wezfurlong.org/wezterm/) or [tmux](https://github.com/tmux/tmux)
- At least one AI provider CLI: [Claude Code](https://github.com/anthropics/claude-code), [Gemini CLI](https://github.com/google-gemini/gemini-cli), [Codex](https://github.com/openai/codex), or [OpenCode](https://opencode.ai)

## Installation

**Note:** Installation differs by provider. Choose the section for your AI tool.

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

### Gemini CLI (via Extension)

```bash
gemini extensions install https://github.com/JsonLee12138/agent-team
```

This installs the `gemini-extension.json` manifest and hooks. The `GEMINI.md` context file is loaded automatically in worktrees.

### OpenCode (via npm Plugin)

```bash
npm install opencode-agent-team
```

Then add to your `opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "plugin": ["opencode-agent-team"]
}
```

Requires `agent-team` binary in PATH (see [Homebrew](#homebrew-macos) or [From Source](#from-source) below).

### Codex

Codex has no plugin or hook system. Install the binary and use the Agent Skill method:

```bash
npx skills add JsonLee12138/agent-team
```

Skills are installed into `.codex/skills/` automatically when creating workers with `--provider codex`. Hook behaviors (brainstorming gate, quality checks) are enforced via prompt conventions in the role's `system.md`.

### Agent Skill (all providers)

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
# Claude Code Plugin
/plugin marketplace update agent-team
# or
claude plugin marketplace update agent-team

# Gemini CLI Extension
gemini extensions update agent-team

# Skill
npx skills add JsonLee12138/agent-team

# Homebrew
brew update && brew upgrade agent-team

# Source
go install github.com/JsonLee12138/agent-team@latest
```

## Quick Start

After installation, you can manage your team entirely through natural language. The AI understands your intent and runs the right commands behind the scenes.

### 1. Create a role

> "Create a frontend developer role for UI implementation."

The AI will guide you through brainstorming the role's scope, goals, and skills, then generate the role package into `.agents/teams/`.

You can also create multiple roles at once:

> "Create a team with frontend developer, QA engineer, and product manager roles."

### 2. List roles

> "Show all available roles."

### 3. Create a worker

> "Create a worker for frontend-dev with claude."

This creates an isolated Git worktree, opens a terminal session, installs required skills, and launches the AI provider.

### 4. Assign a task

> "Assign frontend-dev-001 to implement a responsive navbar."

The AI will run brainstorming first to produce a design doc, then create the task and notify the worker session.

### 5. Check status

> "Show team status."

### 6. Merge and cleanup

> "Merge frontend-dev-001."
>
> "Delete worker frontend-dev-001."

### Install roles from GitHub

> "Search for roles related to react development."
>
> "Install roles from owner/repo."

## Built-in Roles

This repository includes several built-in roles in `.agents/teams/`:

| Role | Description |
|------|-------------|
| `pm` | Product Manager |
| `frontend-architect` | Frontend Architecture |
| `vite-react-dev` | Vite + React Development |
| `uniapp-dev` | UniApp Development |
| `pencil-designer` | Pencil Design Tool Specialist |

> "Create a worker for frontend-architect with claude."

## Built-in Skills

| Skill | Description |
|-------|-------------|
| `role-creator` | Create or update role skill packages interactively |
| `brainstorming` | Turn rough ideas into validated design docs through one-question-at-a-time dialogue before implementation |

## Supported Providers

| Provider | Value | Hook Support | Installation |
|----------|-------|-------------|--------------|
| Claude Code | `claude` (default) | Full (plugin hooks) | Plugin marketplace |
| Gemini CLI | `gemini` | Full (extension hooks) | `gemini extensions install` |
| OpenCode | `opencode` | Full (npm plugin hooks) | npm plugin |
| OpenAI Codex | `codex` | Prompt-driven only | Agent Skill |

**Hook support levels:**
- **Full**: Automatic role injection, brainstorming gate, quality checks, task archiving, idle notification
- **Prompt-driven**: Hook behaviors enforced via role prompts (no automatic interception)

## Advanced

### Role Resolution

When creating a worker or referencing a role, the tool resolves roles with **project-first priority**:

1. **Project**: `.agents/teams/<role-name>/`
2. **Global**: `~/.agents/roles/<role-name>/`

Global roles are referenced **in-place** (not copied to the project). The `worker.yaml` records `role_scope` and `role_path` so subsequent operations (reopen, prompt injection) continue to use the correct source.

### CLI Reference

All commands run inside a Git repository.

#### Role commands

| Command | Description |
|---------|-------------|
| `agent-team role list` | List available roles in `.agents/teams/` |
| `agent-team role create <role-name> --description "..." --system-goal "..." [--force]` | Create or update a role skill package. `--force` skips global duplicate check |

#### Role Repository commands

| Command | Description |
|---------|-------------|
| `agent-team role-repo search <query>` | Search GitHub roles using strict role path contracts |
| `agent-team role-repo add <source> [--role <name>...] [--list] [-g] [-y]` | Discover and install role(s) from `owner/repo` or GitHub URL |
| `agent-team role-repo list [-g]` | List installed repository-managed roles in selected scope |
| `agent-team role-repo remove [roles...] [-g] [-y]` | Remove installed roles and clean lock entries |
| `agent-team role-repo check [-g]` | Check lock entries against remote folder hashes |
| `agent-team role-repo update [-g] [-y]` | Update roles with remote changes |

Accepted remote role path contracts:

- `skills/<role>/references/role.yaml`
- `.agents/teams/<role>/references/role.yaml`

#### Worker commands

| Command | Description |
|---------|-------------|
| `agent-team worker create <role-name> [provider] [--model <model>] [--new-window]` | Create worker, open session, install skills, and launch AI |
| `agent-team worker open <worker-id> [provider] [--model <model>] [--new-window]` | Reopen an existing worker session |
| `agent-team worker assign <worker-id> "<description>" [provider] [--proposal <file>] [--design <file>] [--model <model>] [--new-window]` | Create task change and notify worker |
| `agent-team worker status` | Show workers, roles, running state, skills, and active changes |
| `agent-team worker merge <worker-id>` | Merge `team/<worker-id>` into current branch |
| `agent-team worker delete <worker-id>` | Delete worker worktree, branch, and config |

#### Communication commands

| Command | Description |
|---------|-------------|
| `agent-team reply <worker-id> "<answer>"` | Send `[Main Controller Reply]` message to a worker session |
| `agent-team reply-main "<message>"` | Worker sends `[Worker: <worker-id>]` message to controller session |

### Role Repo Locks

- Project lock: `roles-lock.json`
- Global lock: `~/.agents/.role-lock.json`
- Project install target: `.agents/teams/<role>/`
- Global install target: `~/.agents/roles/<role>/`

### Directory Structure

```
project-root/
├── .agents/
│   └── teams/
│       └── <role-name>/                 <- project roles
│           ├── SKILL.md
│           ├── system.md
│           └── references/role.yaml
├── gemini-extension.json                <- Gemini CLI extension manifest
├── GEMINI.md                            <- Gemini CLI context file
├── hooks/
│   └── hooks.json                       <- Claude + Gemini shared hooks
├── adapters/
│   └── opencode/                        <- OpenCode npm plugin
└── .worktrees/
    └── <worker-id>/
        ├── worker.yaml
        ├── .claude/skills/
        ├── .codex/skills/
        ├── .gemini/skills/
        ├── CLAUDE.md
        ├── AGENTS.md
        ├── GEMINI.md
        └── .tasks/
            └── changes/

~/.agents/
└── roles/
    └── <role-name>/                     <- global roles (referenced in-place)
        ├── SKILL.md
        ├── system.md
        └── references/role.yaml
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `AGENT_TEAM_BACKEND` | Terminal backend: `wezterm` (default) or `tmux` |

## License

MIT
