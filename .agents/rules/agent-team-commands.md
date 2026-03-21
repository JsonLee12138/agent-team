# Agent-Team Commands Rules

## Trigger

Apply this rule whenever a worker session needs project workflow operations such as create, open, assign, archive, reply, or other `agent-team` CLI actions.

## Command Boundary

- MUST use the `agent-team` CLI for worker lifecycle and task lifecycle operations when the repository provides it.
- MUST NOT bypass `agent-team worker open`, `agent-team worker assign`, or `agent-team reply-main` with ad hoc shell commands.
- MUST treat worker bootstrap files and provider prompt files as controller-managed artifacts.

## Generated File Safety

- MUST NOT commit generated worker-local prompt files such as `CLAUDE.md`, `GEMINI.md`, or `AGENTS.md` from a worker worktree.
- MUST NOT commit worker-local metadata such as `.tasks/`, `.claude/`, `.codex/`, `.gemini/`, `.opencode/`, or `worker.yaml` from a worker worktree.
- MUST keep deliverable files in tracked repository paths managed by the assigned change.
