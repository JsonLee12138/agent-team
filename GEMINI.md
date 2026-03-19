<!-- AGENT_TEAM:START -->
# Claude Instructions

Use this file when working in Claude Code on this repository.

- MUST read `.agents/rules/index.md` at task start and load the rule files required by the task.
- MUST call `/compact` whenever any trigger in `.agents/rules/context-management.md` fires.
- MUST keep status updates concise and use `agent-team reply-main` formats from `.agents/rules/communication.md`.
- MUST follow `.agents/rules/task-protocol.md` before reporting completion.
- MUST obey `.agents/rules/worktree.md` for branch and git safety.
- MUST read `.agents/rules/project-commands.md` before running any project command.

## Rules Reference

Load `.agents/rules/index.md` first, then load only the matching rule files:

- `.agents/rules/debugging.md` for bugs, flaky tests, runtime errors, or unexpected behavior
- `.agents/rules/project-commands.md` before running any project command
- `.agents/rules/communication.md` for `reply-main`, blocker escalation, and progress updates
- `.agents/rules/context-management.md` for `/compact` decisions, handoff summaries, and provider-specific context control
- `.agents/rules/task-protocol.md` for task execution, verify, commit, archive, and completion reporting
- `.agents/rules/worktree.md` for branch safety, worktree limits, and ignored path handling

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

# Task management
agent-team task create <worker-id> "<description>"
agent-team task list [--worker <worker-id>]
agent-team task archive <worker-id> <change-name>

# Communication
agent-team reply-main "<message>"
```

## Gemini Notes

- Use a manual summary plus a fresh prompt when context grows too large.
- Keep completion messages short and factual.
- Use `agent-team reply-main` when acting as a worker.
