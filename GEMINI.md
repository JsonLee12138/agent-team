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

## Hooks

The `agent-team hook` subcommands handle lifecycle events automatically:

- **session-start**: Injects role prompt into context files
- **pre-tool-use**: Checks for design documents before coding
- **post-tool-use**: Runs quality checks after edits
- **stop**: Warns about unarchived changes
- **task-completed**: Archives changes and notifies main controller

<!-- Dynamic role content is injected below by the session-start hook -->
