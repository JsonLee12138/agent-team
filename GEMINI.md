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
