# Claude Instructions

Use this file when working in Claude Code on this repository.

- MUST read `.agents/rules/index.md` at task start and load the rule files required by the task.
- MUST call `/compact` whenever any trigger in `.agents/rules/context-management.md` fires.
- MUST keep status updates concise and use `agent-team reply-main` formats from `.agents/rules/communication.md`.
- MUST follow `.agents/rules/task-protocol.md` before reporting completion.
- MUST obey `.agents/rules/worktree.md` for branch and git safety.

<!-- agent-team:rules-start -->
## Rules Reference

Load `.agents/rules/index.md` first, then load only the matching rule files:

- `.agents/rules/debugging.md` for bugs, flaky tests, runtime errors, or unexpected behavior
- `.agents/rules/communication.md` for `reply-main`, blocker escalation, and progress updates
- `.agents/rules/context-management.md` for `/compact` decisions, handoff summaries, and provider-specific context control
- `.agents/rules/task-protocol.md` for task execution, verify, commit, archive, and completion reporting
- `.agents/rules/worktree.md` for branch safety, worktree limits, and ignored path handling
<!-- agent-team:rules-end -->
