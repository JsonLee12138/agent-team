# Hacker News — Show HN Post

> **Guidelines**: Keep title under 80 chars. No clickbait. Technical and factual.
> **Best posting time**: Tuesday–Thursday, 8–10 AM EST
> **Follow-up**: Post a detailed top-level comment immediately after submission

---

## Title

Show HN: Agent-Team – Orchestrate AI coding agents in isolated Git worktrees

## URL

https://github.com/JsonLee12138/agent-team

## Top-Level Comment (post immediately after submission)

Hi HN, I built agent-team because running multiple AI coding agents on the same repo kept producing merge conflicts and context pollution.

The idea is simple: use Git worktrees to give each agent its own isolated copy of the repo, with a dedicated branch and terminal session. A "role" system defines what each agent can and cannot do.

**How it works:**

1. Define a role (e.g., "frontend-architect") with a system prompt, skill dependencies, and scope boundaries
2. Spawn a worker: `agent-team worker create frontend-architect --provider claude`
3. The CLI creates a Git worktree, opens a terminal, and injects the role context
4. Assign tasks via natural language or CLI
5. When done, merge the worker's branch back to main

**Technical decisions worth discussing:**

- **Git worktrees over Docker containers.** Worktrees are lightweight, fast to create/destroy, and natively understand your repo's history. The trade-off is less isolation (shared filesystem), but for AI coding agents this seems like the right level.

- **Brainstorming gates.** A `PreToolUse` hook intercepts file writes. The agent must produce a design document before modifying code. This was the single biggest quality improvement — agents that think before coding produce dramatically better output.

- **5-layer skill resolution.** Skills are searched hierarchically: plugin → project → global → user home → remote GitHub. This means a team can share skills via a repo while individuals can override locally.

- **Provider-agnostic roles.** Same YAML role file works across Claude Code, Gemini CLI, OpenCode, and OpenAI Codex. The hook system adapts to each provider's event model.

**What ships out of the box:** 11 pre-built roles (PM, frontend architect, React dev, UI designer, QA, mobile dev, etc.), 7 lifecycle hooks, and a companion dashboard (Remix app) for browsing the role library.

**Stack:** Go CLI, YAML configs, shell-based hooks. MIT licensed.

I'd love feedback on the architecture, especially from anyone who's tried other approaches to multi-agent coordination.

---

## Anticipated Questions & Prepared Responses

### "How is this different from just opening multiple terminal tabs?"

The isolation is the key difference. Multiple tabs still share the same working directory and branch. When two agents edit the same file, you get conflicts. With worktrees, each agent has its own filesystem copy and branch — changes are independent until you explicitly merge.

Beyond that, the role system enforces scope boundaries. A frontend agent literally cannot modify backend files because its scope definition excludes them. And lifecycle hooks add quality gates that raw terminals don't have.

### "Why Git worktrees instead of Docker?"

Speed and simplicity. Creating a worktree takes milliseconds vs seconds for a container. Worktrees share the .git directory so they're space-efficient. And worktrees understand your repo's history natively — no volume mounting needed.

The trade-off is less sandboxing. If you need process-level isolation (e.g., untrusted agents), Docker is better. But for AI coding assistants from major providers, worktree-level isolation is the right balance.

### "Does this actually improve output quality?"

The brainstorming gate is the biggest factor. We hook into file-write events and require a design document before any code modification. Anecdotally, this eliminates the "agent wrote 500 lines without thinking" failure mode.

The scope boundaries also help — agents that know their lane tend to produce more focused, higher-quality code than agents with no constraints.

### "Why Go?"

Fast single-binary distribution, cross-platform compilation, and excellent CLI tooling (cobra). The CLI needs to be fast because it runs as a hook handler — every file write triggers a hook that must respond in under 5 seconds.
