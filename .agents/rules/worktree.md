# Worktree Rules

## Trigger

Apply this rule for any git command, branch action, file placement decision, or task work inside a worker worktree.

## Branch and Worktree Safety

- MUST work only inside the assigned worktree.
- MUST keep all task commits on the current `team/<worker-id>` branch.
- NEVER run `git checkout`, `git switch`, `git merge`, or `git rebase` inside the worker worktree.
- NEVER modify another worker's worktree or branch from the current session.

## File Placement

- MUST keep deliverables in tracked repository paths.
- MUST NOT place task outputs in ignored locations such as `.claude/`, `.codex/`, `.gemini/`, `.opencode/`, `.tasks/`, or `worker.yaml`.
- ALWAYS check whether new files belong to the task scope before creating them.

## Staging and Commit Scope

- MUST inspect `git status` before staging changes.
- MUST stage only files required for the assigned change.
- NEVER use blanket staging commands that may capture unrelated work.
