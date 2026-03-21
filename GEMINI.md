<!-- AGENT_TEAM:START -->
# Claude Instructions

Use this file when working in Claude Code on this repository.

- MUST read `.agent-team/rules/index.md` at task start and load only the rule files required by the task.
- MUST follow `.agent-team/rules/context-management.md` for context-cleanup and index-first recovery whenever context drifts, phases change, or work resumes.
- MUST keep status updates concise.
- MUST obey `.agent-team/rules/worktree.md` for branch and git safety.
- MUST read `.agent-team/rules/index.md` at task start and then open only the relevant files under `.agent-team/rules/core/` and `.agent-team/rules/project/`.
- MUST follow `.agent-team/rules/core/context-management.md` for context-cleanup and index-first recovery whenever context drifts, phases change, or work resumes.
- MUST keep status updates concise.
- MUST obey `.agent-team/rules/core/worktree.md` for branch and git safety.
- MUST follow `.agent-team/rules/core/agent-team-commands.md` for worker lifecycle and generated-file boundaries.
- MUST follow `.agent-team/rules/core/merge-workflow.md` for controller-side rebase and merge sequencing.

## Rules Reference

Load `.agent-team/rules/index.md` first, then load only the matching rule files:

- `.agent-team/rules/core/debugging.md` for bugs, flaky tests, runtime errors, or unexpected behavior
- `.agent-team/rules/core/agent-team-commands.md` for agent-team CLI boundaries and worker lifecycle operations
- `.agent-team/rules/core/merge-workflow.md` for controller-side rebase, merge ordering, and generated file safety
- `.agent-team/rules/core/context-management.md` for context-cleanup triggers, session reset boundaries, and index-first file recovery
- `.agent-team/rules/core/worktree.md` for branch safety, worktree limits, and ignored path handling
- Read the relevant files under `.agent-team/rules/project/` before running project-specific commands or workflows

<!-- AGENT_TEAM:END -->

# Agent Team

This project uses [agent-team](https://github.com/JsonLee12138/agent-team) for multi-agent development workflows with git worktrees.

## Overview

`agent-team` manages AI coding agents using a **Role** (skill package) + **Worker** (instance) model:

- **Roles** define skills, system prompts, and quality checks for a specific function.
- **Workers** are isolated git worktree instances assigned to a role, each with its own branch.

## Commands

```bash
# Team management
agent-team worker create <role> [--provider gemini]
agent-team worker status
agent-team worker open <worker-id> [--provider gemini]

# Assignment
agent-team worker assign <worker-id> "<description>" [--proposal <file>] [--design <file>]

# Communication
agent-team reply-main "<message>"
```

## Gemini Notes

- Use context-cleanup plus a fresh, index-first file read when session context drifts.
- Keep completion messages short and factual.
- Use `agent-team reply-main` when acting as a worker.
