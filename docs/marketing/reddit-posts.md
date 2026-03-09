# Reddit Posts — agent-team Launch

---

## Post 1: r/ClaudeAI

### Title
I built a CLI to run multiple Claude agents in parallel without merge conflicts — agent-team

### Body

I've been running into the same problem over and over: I tell Claude to work on the frontend, then spin up another session for the backend, and they step all over each other's files.

So I built **agent-team** — an open-source CLI that creates isolated Git worktrees for each AI agent. Each agent gets:

- Its own branch (no merge conflicts)
- Its own terminal session
- A pre-defined "role" with skills, scope boundaries, and system prompts
- Quality gates (brainstorming gate forces design docs before code)

**What makes it Claude-specific:**

- Ships as a **Claude Plugin** (`/plugin marketplace add JsonLee12138/agent-team`)
- 7 lifecycle hooks integrate directly with Claude Code's hook system
- The brainstorming gate hooks into `Write` and `Edit` tool calls — Claude literally can't write code until it justifies the approach
- Skills auto-inject into Claude's context via `SessionStart` hook

**11 pre-built roles** out of the box: PM, Frontend Architect, Vite+React Dev, UI Designer, QA, Growth Marketer, etc. Install community roles with one command.

**Quick start:**

```bash
npx skills add JsonLee12138/agent-team -a claude -y
```

Then just tell Claude: "Create a worker for frontend-architect."

It creates a worktree, opens a terminal, injects the role prompt, and starts working.

Repo: https://github.com/JsonLee12138/agent-team

Happy to answer questions. What roles would you want to see?

---

## Post 2: r/LocalLLaMA

### Title
Open-source multi-agent orchestrator that works across Claude, Gemini, OpenCode & Codex — Git worktree isolation for parallel AI dev

### Body

I know this sub loves provider-agnostic tooling, so I wanted to share **agent-team** — an open-source CLI (Go, MIT) that orchestrates multiple AI coding agents regardless of which provider you use.

**The core idea:** Each agent runs in an isolated Git worktree with its own branch and terminal. No merge conflicts, no context bleeding. You define "roles" (skill packages) once, then spawn "workers" on any provider.

**Supported providers:**
- Claude Code (full plugin + hooks)
- Gemini CLI (full extension + hooks)
- OpenCode (full NPM plugin)
- OpenAI Codex (prompt-only, no hooks yet)

**Why this matters for provider diversity:**

You can assign the same task to different providers and compare outputs. Frontend task to Claude, backend to Gemini, tests to Codex — they all work in isolation and merge cleanly.

**Key features:**
- 11 pre-built roles (PM, frontend, backend, QA, designer, etc.)
- 5-layer skill resolution (project → global → remote)
- 7 lifecycle hooks for quality gates
- Brainstorming gate: agents must design before coding
- Task archiving with full Git audit trail
- Natural language task assignment

**Install:**

```bash
brew tap JsonLee12138/agent-team && brew install agent-team
# or
go install github.com/JsonLee12138/agent-team@latest
```

Repo: https://github.com/JsonLee12138/agent-team

The role system is extensible — you can create custom roles or pull from community repos. Would love feedback from folks running local models on whether the OpenCode integration works well for your setup.

---

## Post 3: r/programming

### Title
Show r/programming: agent-team — orchestrate AI coding agents in isolated Git worktrees (Go, MIT, multi-provider)

### Body

**Problem:** When you run multiple AI coding agents on the same repo, they create merge conflicts, overwrite each other's work, and have no concept of "staying in their lane."

**Solution:** [agent-team](https://github.com/JsonLee12138/agent-team) uses Git worktrees to give each agent an isolated copy of the repo with its own branch and terminal session.

**Architecture:**

```
project-root/
├── .agents/teams/        ← Role definitions (reusable skill packages)
├── .worktrees/           ← Isolated worker workspaces
│   ├── frontend-001/     ← Claude working on UI
│   ├── backend-001/      ← Gemini working on API
│   └── qa-001/           ← Codex writing tests
├── roles-lock.json       ← Version locking for reproducibility
└── hooks/                ← Lifecycle hooks (quality gates)
```

**Design decisions I think are interesting:**

1. **Roles are skill packages, not just prompts.** A role includes a system prompt, dependency skills, scope boundaries (what the agent CAN and CANNOT do), and tool configurations. This prevents the "agent does everything poorly" problem.

2. **Brainstorming gates.** A `PreToolUse` hook intercepts `Write` and `Edit` tool calls. The agent must produce a design document before it can modify any file. This dramatically improves output quality.

3. **5-layer skill resolution.** Skills are searched: plugin built-in → project → global Claude skills → global Codex skills → user home. Falls back to remote GitHub download if nothing matches.

4. **Provider-agnostic roles.** Same role definition works on Claude Code, Gemini CLI, OpenCode, and OpenAI Codex. Write once, run anywhere (for AI agents).

**Tech stack:** Go CLI, YAML configs, shell hooks. The companion dashboard (role-hub) is a Remix fullstack app.

**Not trying to sell anything** — this is MIT licensed and the whole thing is on GitHub. I'm genuinely curious how other teams handle multi-agent coordination, because every solution I've seen either uses Docker containers (heavy) or hopes for the best (chaos).

Repo: https://github.com/JsonLee12138/agent-team

---

## Engagement Guidelines

### r/ClaudeAI
- Monitor for questions about Claude-specific features
- Offer to help with custom role creation
- Share code snippets for hook configuration

### r/LocalLLaMA
- Emphasize provider flexibility and local model compatibility
- Ask for feedback on OpenCode integration
- Discuss future plans for more providers

### r/programming
- Focus on architecture and design decisions
- Be ready to discuss trade-offs (worktrees vs containers vs namespaces)
- Avoid marketing language — this sub values technical depth
