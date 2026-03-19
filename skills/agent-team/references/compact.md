# Compact Command Reference

## Purpose

`agent-team compact` manually injects Claude's built-in `/compact` command into a recorded Claude session pane.

This command is for stable manual delivery only:
- it sends `/compact` to the resolved pane
- it does not wait for completion
- it does not inspect Claude's compact result

## Command

```bash
agent-team compact [--pane-id <id>] [--worker <worker-id>] [--to main] [--message "<text>"]
```

## Target Resolution Order

1. `--pane-id <id>`
   - Direct pane override.

2. Current worker worktree
   - If run inside a worker worktree, use the current `worker.yaml` `pane_id`.

3. `--to main`
   - In a worker worktree: use `controller_pane_id`
   - In repo root: use the project main pane config
   - If the project main pane config is missing: fall back once to `WEZTERM_PANE` / `TMUX_PANE` and persist it

4. Repo root without `--to main`
   - Require `--worker <worker-id>`
   - Target that worker's `pane_id`

## Examples

```bash
# From a worker worktree, compact the current worker session
agent-team compact

# From repo root, compact a specific worker
agent-team compact --worker dev-001

# From a worker worktree, compact the controller/main session
agent-team compact --to main

# Send an annotated compact instruction
agent-team compact --message "Focus on current task and next step"

# Escape hatch: send directly to a pane
agent-team compact --pane-id 123456
```

## Notes

- Backend preference is WezTerm first, tmux second.
- Claude workers still MUST follow `.agents/rules/context-management.md` and compact when rule triggers fire.
- The controller/main pane is recorded at project level so `--to main` can target the main Claude session reliably.
