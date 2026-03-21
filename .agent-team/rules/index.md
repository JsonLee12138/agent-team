# Rules Index

Read this file first. It is the single entry point for project rules.

## Core Rules

- `core/debugging.md`: use for bugs, flaky tests, runtime errors, build failures, and unexpected behavior.
- `core/agent-team-commands.md`: use for worker lifecycle, task lifecycle, and agent-team CLI boundaries.
- `core/merge-workflow.md`: use for controller-side rebase, synchronization, and merge sequencing.
- `core/context-management.md`: use for context-cleanup, index-first recovery, and resume rules.
- `core/worktree.md`: use for branch safety, worktree limits, and file placement.

## Project Rules

- Read the relevant files under `project/` before running project-specific commands or workflows.
- The `project/` directory is AI-generated during `agent-team init` and refreshed by `agent-team rules sync`.
- Project rules must stay split by single responsibility. Do not collapse them back into this index.
