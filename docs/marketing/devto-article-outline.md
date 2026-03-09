# Dev.to Article Outline

> **Title**: How I Orchestrate a Team of AI Coding Agents Without Chaos
> **Tags**: `ai`, `devtools`, `opensource`, `productivity`, `tutorial`
> **Series**: (standalone, expandable to series)
> **Estimated length**: 2,000–2,500 words
> **Cover image**: Terminal screenshot showing 3 agent workers in split panes

---

## Article Structure

### Hook (150 words)
- Open with the pain point: "I had three AI agents working on my project. Two of them edited the same file. One overwrote the other's changes. Nobody noticed until the tests broke."
- Transition: "I realized the problem isn't AI quality — it's coordination. So I built a tool to solve it."
- Introduce agent-team in one line

### Section 1: The Problem with Multi-Agent Development (300 words)

**Key points:**
- AI coding agents are powerful individually but chaotic in groups
- Common failure modes:
  - Merge conflicts from concurrent edits
  - Context pollution (Agent B inherits Agent A's mistakes)
  - Scope creep (every agent tries to do everything)
  - No quality gates (agents ship untested, unreviewed code)
- Why "just use branches" doesn't work (agents don't coordinate branch strategy)
- The missing piece: orchestration layer between human and agents

### Section 2: The Role + Worker Model (400 words)

**Key points:**
- Explain the architecture with a diagram:
  ```
  Human (Main Controller)
       │
       ├── Role: frontend-architect
       │   └── Worker: frontend-001 (Claude, worktree, branch)
       │
       ├── Role: backend-dev
       │   └── Worker: backend-001 (Gemini, worktree, branch)
       │
       └── Role: qa-engineer
           └── Worker: qa-001 (OpenCode, worktree, branch)
  ```
- Roles = reusable templates (system prompt + skills + scope)
- Workers = isolated instances (Git worktree + branch + terminal)
- Why scope boundaries matter (what agent CAN vs CANNOT do)
- Show a role YAML example

### Section 3: Getting Started — 5-Minute Tutorial (500 words)

**Step-by-step walkthrough:**
1. Install agent-team (Homebrew / Go / AI-native)
2. Initialize a project: `agent-team init`
3. List available roles: `agent-team role list`
4. Create a worker: show the command and explain what happens (worktree creation, branch, terminal, prompt injection)
5. Assign a task: natural language or CLI
6. Monitor progress: `agent-team worker status`
7. Merge and cleanup: show the merge command

**Include:** Terminal screenshots or code blocks at each step

### Section 4: The Brainstorming Gate — Why It Changes Everything (350 words)

**Key points:**
- The #1 quality problem with AI agents: they code before they think
- How the brainstorming gate works:
  - `PreToolUse` hook intercepts Write/Edit events
  - Agent must produce a design document first
  - Only after design approval can code be written
- Before vs after comparison (conceptual)
- Why this is better than post-hoc code review
- How to configure or disable the gate

### Section 5: Multi-Provider Flexibility (300 words)

**Key points:**
- Same role works on Claude, Gemini, OpenCode, Codex
- Why this matters:
  - Different strengths per provider (Claude for reasoning, Gemini for speed, etc.)
  - No vendor lock-in
  - Cost optimization (use cheaper models for simpler tasks)
- Provider-specific integrations (Claude Plugin, Gemini Extension)
- How the hook system adapts to each provider

### Section 6: Building Custom Roles (300 words)

**Key points:**
- Walk through creating a custom role
- Show the YAML structure (name, description, system prompt, skills, scope)
- Skills dependency system (5-layer search)
- Sharing roles via GitHub repos (`agent-team role-repo add`)
- Community role contribution model

### Conclusion & CTA (200 words)

- Recap: multi-agent development needs orchestration, not just more agents
- agent-team provides isolation (worktrees), quality (gates), and flexibility (multi-provider)
- Call to action:
  - Star the repo: github.com/JsonLee12138/agent-team
  - Try it: 5-minute install
  - Contribute: create and share roles
  - Join the discussion: link to GitHub Discussions

---

## Content Assets Needed

1. **Cover image**: Split-pane terminal with 3 agent workers active
2. **Architecture diagram**: Role + Worker model (Mermaid or Excalidraw)
3. **Code blocks**: Role YAML, CLI commands, hook configuration
4. **Before/after**: Output comparison with and without brainstorming gate
5. **Provider comparison table**: Feature matrix across 4 providers

## SEO Notes

- Primary keyword: "multi-agent AI development"
- Secondary: "AI coding agent orchestration", "Git worktree AI agents", "Claude Code multi-agent"
- Internal links: Link to GitHub repo, role-hub dashboard
- Cross-promotion: Reference the Twitter thread and Reddit posts
