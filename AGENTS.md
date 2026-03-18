# Codex Instructions

Use this file when working in Codex CLI on this repository.

- MUST read `.agents/rules/index.md` at task start and load the rule files required by the task.
- MUST keep `update_plan` current for multi-step work and send short preambles before grouped tool calls.
- MUST use `apply_patch` for manual file edits and run targeted verification before the final handoff.
- MUST follow manual compaction from `.agents/rules/context-management.md` because Codex has no native `/compact`.
- MUST use `agent-team reply-main` exactly as defined in `.agents/rules/communication.md` and `.agents/rules/task-protocol.md`.

<!-- agent-team:rules-start -->
## Rules Reference

Load `.agents/rules/index.md` first, then load only the matching rule files:

- `.agents/rules/debugging.md` for bugs, flaky tests, runtime errors, or unexpected behavior
- `.agents/rules/communication.md` for `reply-main`, blocker escalation, and progress updates
- `.agents/rules/context-management.md` when context grows, handoff is needed, or provider state is degrading
- `.agents/rules/task-protocol.md` for task execution, verify, commit, archive, and completion reporting
- `.agents/rules/worktree.md` for branch safety, worktree limits, and ignored path handling
<!-- agent-team:rules-end -->
