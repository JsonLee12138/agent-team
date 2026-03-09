# X/Twitter Launch Thread — agent-team

> **Target**: Developer audience, AI-native builders, tech leads
> **Hashtags**: #AgentTeam #AIAgents #DevTools #OpenSource #MultiAgent #ClaudeCode #GeminiCLI
> **Optimal posting time**: Tuesday–Thursday, 9–11 AM PST

---

## Thread

**1/12**
We just open-sourced agent-team — a CLI that lets you orchestrate multiple AI agents in isolated Git worktrees.

No merge conflicts. No prompt chaos. Just surgical precision.

🔗 github.com/JsonLee12138/agent-team

🧵 Here's why this changes everything ↓

---

**2/12**
The problem: you give Claude a task, then Gemini another, then Codex a third.

They all edit the same files. Branches collide. Context bleeds. You spend more time fixing agent mess than writing code.

Sound familiar?

---

**3/12**
agent-team introduces the **Role + Worker** model:

- **Role** = reusable skill package (system prompt + tools + scope)
- **Worker** = isolated runtime (own branch + worktree + terminal)

Each agent gets its own sandbox. Zero interference.

---

**4/12**
How it works:

```
> "Create a frontend-architect worker using Claude"
```

That's it. agent-team:
1. Creates a Git worktree
2. Opens a dedicated terminal
3. Injects the role's skills & prompts
4. Assigns the worker its own branch

---

**5/12**
11 pre-built roles ship out of the box:

- Product Manager
- Frontend Architect
- Vite + React Dev
- UI Designer (Pencil MCP)
- QA & Observability
- Growth Marketer
- UniApp Cross-Platform Dev
- ...and more

Install one: `agent-team role-repo add JsonLee12138/agent-team`

---

**6/12**
Works with 4 AI providers — no vendor lock-in:

| Provider | Support |
|----------|---------|
| Claude Code | Full plugin ✅ |
| Gemini CLI | Full extension ✅ |
| OpenCode | Full plugin ✅ |
| OpenAI Codex | Prompt-only ⚠️ |

Same role, any provider. Your choice.

---

**7/12**
The killer feature: **Brainstorming Gates**

Before any agent writes code, it MUST produce a design doc.

No more "AI generated 1,000 lines without thinking."

A quality gate hooks into Write/Edit events and enforces design-first development.

---

**8/12**
7 lifecycle hooks power the quality system:

1. SessionStart → role prompt injection
2. PreToolUse → brainstorming gate
3. PostToolUse → quality checks
4. TaskCompleted → auto-archive
5. Stop → unmerged change warnings
6. SubagentStart → context propagation
7. TeammateIdle → collaboration signals

---

**9/12**
Installation? AI-native:

```bash
npx skills add JsonLee12138/agent-team -a claude -y
```

Then tell your agent: "Install agent-team and initialize the project."

Your AI sets itself up. That's the whole point.

---

**10/12**
Task management is Git-backed:

- Every agent output tied to a changeset
- Task archiving with completion tracking
- Merge & cleanup in one command
- Full audit trail: who did what, when, why

```
> "Merge frontend-architect-001 and delete the worker"
```

---

**11/12**
What we're building next:

- Role marketplace with community contributions
- Role-Hub dashboard (already in beta — Remix fullstack)
- Cross-agent collaboration protocols
- More providers (Cursor, Windsurf)

Star ⭐ the repo if this solves a problem for you.

---

**12/12**
agent-team is MIT licensed and built with Go.

Get started in 60 seconds:
🔗 github.com/JsonLee12138/agent-team

Or install via Homebrew:
```bash
brew tap JsonLee12138/agent-team && brew install agent-team
```

Questions? DM me or open an issue.

#OpenSource #AIAgents #DevTools

---

## Engagement Strategy

1. **Quote-tweet** tweet 7 (brainstorming gates) with a short video demo
2. **Pin** the thread for 48 hours
3. **Reply** to comments within 30 minutes during the first 6 hours
4. **Cross-post** key tweets to LinkedIn with expanded context
5. **Schedule follow-up** thread 3 days later with user testimonials or demo GIFs
