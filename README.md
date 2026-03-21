# agent-team

English | [中文](./README.zh.md)

**Orchestrate your AI workforce with surgical precision.** 🚀

`agent-team` is a multi-agent development manager that uses a **Role + Worker** model to run AI agents in isolated Git worktrees and dedicated terminal sessions.

- **🎭 Role**: Reusable skill packages (`.agent-team/teams/`) defining goals, prompts, and tools.
- **🛠️ Worker**: Isolated runtime instances (`.worktrees/`) with their own branch and session.

---

## Table of Contents

- [Installation](#️-installation)
  - [AI-Native (Recommended)](#-ai-native-recommended)
  - [Manual Installation](#-manual-installation)
- [Upgrade](#-upgrade)
- [Quick Start](#-quick-start)
- [The Toolkit](#-the-toolkit)
- [Supported Providers](#-supported-providers)
- [Advanced Usage](#️-advanced-usage)
  - [CLI Reference](#cli-reference)
  - [Directory Structure](#directory-structure)
  - [Environment Variables](#environment-variables)
- [License](#-license)

---

## 🛠️ Installation

### 🤖 AI-Native (Recommended)
Let your AI agent set itself up. It only takes two steps:

1. **Install the Skill**:
   ```bash
   npx skills add JsonLee12138/agent-team -a <platform> -y
   ```
   Replace `<platform>` with your provider: `claude`, `gemini`, `opencode`, or `codex`.
2. **Tell your Agent**:
   > "Install agent-team and initialize the project."

---

### 📦 Manual Installation

| Method | Command |
| :--- | :--- |
| **Homebrew** | `brew tap JsonLee12138/agent-team && brew install agent-team` |
| **Go Install** | `go install github.com/JsonLee12138/agent-team@latest` |
| **Claude Plugin** | `/plugin marketplace add JsonLee12138/agent-team` |
| **Gemini Ext** | `gemini extensions install https://github.com/JsonLee12138/agent-team` |
| **OpenCode Plugin** | `{ "plugin": ["opencode-agent-team"] }` in `~/.config/opencode/opencode.json` |

---

## 🔄 Upgrade

| Method | Command |
| :--- | :--- |
| **Claude Plugin** | `/plugin marketplace update agent-team` |
| **Skill** | `npx skills add JsonLee12138/agent-team -a '*' -y` |
| **Homebrew** | `brew update && brew upgrade agent-team` |
| **Go Install** | `go install github.com/JsonLee12138/agent-team@latest` |

---

## 🚀 Quick Start

Manage your entire team through natural language. Your AI agent will handle the commands for you.

### 1. Define your Team
> "Create a **frontend-architect** role to manage our UI infrastructure."

### 2. Install Curated Roles (Optional)
Pull in expert roles from this repository:
```bash
agent-team role-repo add JsonLee12138/agent-team
```

### 3. Spawn a Worker
> "Create a worker for **frontend-architect** using Claude."
*This opens a new terminal window with an isolated worktree.*

### 4. Assign & Brainstorm
> "Assign **frontend-architect-001** to design the new auth flow."
*The agent will brainstorm a design doc before touching any code.*

### 5. Merge & Cleanup
> "Merge **frontend-architect-001** and delete the worker."

---

## 🧰 The Toolkit

### Built-in Roles
Available in `.agent-team/teams/`:
- `pm`: Product Manager & Requirement Shaping.
- `frontend-architect`: High-level UI/UX structure.
- `vite-react-dev`: Specialized for Vite + React.
- `pencil-designer`: UI design tool specialist.

### Built-in Skills
- `task-orchestrator`: Task lifecycle entry.
- `workflow-orchestrator`: Governance-only workflow plan entry.
- `worker-dispatch`: Controller-side worker open/reply entry.
- `worker-recovery`: Worker-first resume entry.
- `worker-reply-main`: Worker -> main reporting entry.
- `context-cleanup`: Session cleanup + index-first file re-anchoring.
- `task-inspector`: Read-only task inspection.
- `role-repo-manager`: Role source management.
- `catalog-browser`: Read-only catalog browsing.
- `project-bootstrap`: `init` / `migrate` entry.
- `rules-maintenance`: `rules sync` entry.
- `skill-maintenance`: Skill cache maintenance.
- `worker-inspector`: Read-only worker status.
- `role-browser`: Read-only local role browsing.
- `role-creator`: Interactively build new agent roles.
- `brainstorming`: Validates ideas via dialogue before implementation, including planning-target selection (`roadmap` / `milestone` / `phase` / `task` / `generic topic`) and explicit save-destination choices (`docs/brainstorming/`, target-object directory, custom path, or skip saving).
- `agent-team`: Legacy compatibility shell that now routes to the dedicated scenario skills.

---

## 🤖 Supported Providers

| Provider | CLI Value | Integration |
| :--- | :--- | :--- |
| **Claude Code** | `claude` | Plugin |
| **Gemini CLI** | `gemini` | Extension |
| **OpenCode** | `opencode` | NPM Plugin |
| **OpenAI Codex** | `codex` | Prompt-only |

---

## ⚙️ Advanced Usage

<details>
<summary><b>📖 CLI Reference</b></summary>

### Role Management
- `agent-team role list`: Show local roles.
- `agent-team role create <name>`: Create a new role package (`SKILL.md`, `references/role.yaml`, `system.md`) under `skills/`, `.agent-team/teams/`, or a custom target path.
- `agent-team role-repo add <owner/repo>`: Install roles from GitHub.

### Worker Operations
- `agent-team worker create <role> [--provider <provider>] [--model <model>]`: Prepare a new worker (does not start a session).
- `agent-team worker open <worker-id> [--provider <provider>] [--model <model>] [--new-window]`: Start or reopen a worker session.
- `agent-team worker close <worker-id>`: Close a worker session without deleting the worker.
- `agent-team worker status`: View active workers and tasks.
- `agent-team worker assign <id> "<task>"`: Dispatch work.
- `agent-team worker merge <id>`: Sync worker changes back (does not close the session).
- `agent-team worker delete <id>`: Remove a worker and its worktree.

### Communication
- `agent-team reply <id> "<msg>"`: Send message to worker.
- `agent-team reply-main "<msg>"`: Worker talks back to main.

### Planning Artifacts
- `agent-team planning create --kind <roadmap|milestone|phase> "<title>"`: Create a planning artifact.
- `agent-team planning list [--kind <kind>] [--lifecycle <planning|archived|deprecated>]`: List planning artifacts.
- `agent-team planning show <id>`: Show a planning artifact and reference checks.
- `agent-team planning move <id> --to <planning|archived|deprecated>`: Move a planning artifact across lifecycle directories.

### Task Artifacts
Every active task package contains three standard artifacts under `.agent-team/task/<task-id>/`:
- `task.yaml`: lifecycle metadata and status.
- `context.md`: background, scope, constraints, and design input.
- `verification.md`: acceptance contract, test scope boundary, performed checks, and final verification result.

`verification.md` is created automatically with `agent-team task create`. Its default template keeps `E2E Required: no`, `Verified By: qa`, and a pending result so QA or human acceptance can complete the record later.

Lifecycle summary:
- `task done` now moves a task from `assigned` to `verifying`.
- `task archive` reads `verification.md` and only archives tasks whose `## Result` passes the gate.
- `task deprecated` moves unfinished or abandoned work into `.agent-team/deprecated/task/<task-id>/` while preserving the task package.

</details>

<details>
<summary><b>📂 Directory Structure</b></summary>

```
project-root/
├── .agent-team/task/                  <- Active task packages
├── .agent-team/planning/roadmaps/     <- Active roadmap artifacts
├── .agent-team/planning/milestones/   <- Active milestone artifacts
├── .agent-team/planning/phases/       <- Active phase artifacts
├── .agent-team/archive/task/          <- Archived task packages
├── .agent-team/archive/roadmaps/      <- Archived roadmap artifacts
├── .agent-team/archive/milestones/    <- Archived milestone artifacts
├── .agent-team/archive/phases/        <- Archived phase artifacts
├── .agent-team/deprecated/task/       <- Deprecated task packages
├── .agent-team/deprecated/roadmaps/   <- Deprecated roadmap artifacts
├── .agent-team/deprecated/milestones/ <- Deprecated milestone artifacts
├── .agent-team/deprecated/phases/     <- Deprecated phase artifacts
├── .agent-team/teams/                 <- Project-specific roles
├── .worktrees/                        <- Isolated worker workspaces
├── roles-lock.json                    <- Remote role version locking
└── gemini-extension.json              <- Extension manifest
```

</details>

<details>
<summary><b>🌐 Environment Variables</b></summary>

| Variable | Default | Description |
| :--- | :--- | :--- |
| `AGENT_TEAM_BACKEND` | `wezterm` | Terminal: `wezterm` or `tmux`. |
| `AGENT_TEAM_ROLE_HUB_URL` | `https://...` | Ingest endpoint for analytics. |
| `AGENT_TEAM_ROLE_HUB_DEBUG` | `0` | Wait for ingest if set to `1`. |

</details>

---

## 📄 License

MIT © [JsonLee](https://github.com/JsonLee12138)
