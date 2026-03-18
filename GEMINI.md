# Agent Team

This project uses [agent-team](https://github.com/JsonLee12138/agent-team) for multi-agent development workflows with git worktrees.

## Overview

`agent-team` manages AI coding agents using a **Role** (skill package) + **Worker** (instance) model:

- **Roles** define skills, system prompts, and quality checks for a specific function (e.g., frontend-dev, backend-dev)
- **Workers** are isolated git worktree instances assigned to a role, each with its own branch

## Commands

```bash
# Team management
agent-team worker create <role> [--provider gemini]   # Create a worker
agent-team worker list                                 # List all workers
agent-team worker open <worker-id> [--provider gemini] # Open worker session

# Task management
agent-team task create <worker-id> <name> <description>  # Create a change
agent-team task list <worker-id>                          # List changes
agent-team task archive <worker-id> <change-name>         # Archive completed change

# Communication
agent-team reply-main "<message>"                      # Send message to main controller
```

<!-- Dynamic role content is injected below by worker open -->

## Gemini Notes

- MUST read `.agents/rules/index.md` at task start and load the rule files required by the task.
- MUST use a manual summary plus a fresh prompt when `.agents/rules/context-management.md` says to compact, because Gemini CLI has no native `/compact`.
- MUST use `agent-team reply-main` formats from `.agents/rules/communication.md`.
- MUST complete the commit -> archive -> `reply-main` chain from `.agents/rules/task-protocol.md` before reporting done.

<!-- agent-team:rules-start -->
## Rules Reference

Load `.agents/rules/index.md` first, then load only the matching rule files:

- `.agents/rules/debugging.md` for bugs, flaky tests, runtime errors, or unexpected behavior
- `.agents/rules/communication.md` for `reply-main`, blocker escalation, and progress updates
- `.agents/rules/context-management.md` for manual compaction, handoff summaries, and provider-specific context control
- `.agents/rules/task-protocol.md` for task execution, verify, commit, archive, and completion reporting
- `.agents/rules/worktree.md` for branch safety, worktree limits, and ignored path handling
<!-- agent-team:rules-end -->
