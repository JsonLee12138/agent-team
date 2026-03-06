# agent-team

English | [中文](./README.zh.md)

**Orchestrate your AI workforce with surgical precision.** 🚀

`agent-team` is a multi-agent development manager that uses a **Role + Worker** model to run AI agents in isolated Git worktrees and dedicated terminal sessions.

- **🎭 Role**: Reusable skill packages (`.agents/teams/`) defining goals, prompts, and tools.
- **🛠️ Worker**: Isolated runtime instances (`.worktrees/`) with their own branch and session.

---

## 🛠️ Installation

### 🤖 AI-Native (Recommended)
Let your AI agent set itself up. It only takes two steps:

1. **Install the Skill**:
   ```bash
   npx skills add JsonLee12138/agent-team -a '*' -y
   ```
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
Available in `.agents/teams/`:
- `pm`: Product Manager & Requirement Shaping.
- `frontend-architect`: High-level UI/UX structure.
- `vite-react-dev`: Specialized for Vite + React.
- `pencil-designer`: UI design tool specialist.

### Built-in Skills
- `role-creator`: Interactively build new agent roles.
- `brainstorming`: Validates ideas via dialogue before implementation.

---

## 🤖 Supported Providers

| Provider | CLI Value | Hook Support |
| :--- | :--- | :--- |
| **Claude Code** | `claude` | ✅ Full (Plugin) |
| **Gemini CLI** | `gemini` | ✅ Full (Extension) |
| **OpenCode** | `opencode` | ✅ Full (NPM Plugin) |
| **OpenAI Codex** | `codex` | ⚠️ Prompt-only |

---

## ⚙️ Advanced Usage

<details>
<summary><b>📖 CLI Reference</b></summary>

### Role Management
- `agent-team role list`: Show local roles.
- `agent-team role create <name>`: Create a new role package.
- `agent-team role-repo add <owner/repo>`: Install roles from GitHub.

### Worker Operations
- `agent-team worker create <role>`: Spin up a new worker.
- `agent-team worker status`: View active workers and tasks.
- `agent-team worker assign <id> "<task>"`: Dispatch work.
- `agent-team worker merge <id>`: Sync worker changes back.

### Communication
- `agent-team reply <id> "<msg>"`: Send message to worker.
- `agent-team reply-main "<msg>"`: Worker talks back to main.

</details>

<details>
<summary><b>📂 Directory Structure</b></summary>

```
project-root/
├── .agents/teams/        <- Project-specific roles
├── .worktrees/           <- Isolated worker workspaces
├── roles-lock.json       <- Remote role version locking
├── gemini-extension.json <- Extension manifest
└── hooks/                <- Shared lifecycle hooks
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
