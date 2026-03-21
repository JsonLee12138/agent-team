<!-- AGENT_TEAM:START -->
# Claude Instructions

Use this file when working in Claude Code on this repository.

- MUST read `.agents/rules/index.md` at task start and load the rule files required by the task.
- MUST call `/compact` whenever any trigger in `.agents/rules/context-management.md` fires.
- MUST keep status updates concise.
- MUST obey `.agents/rules/worktree.md` for branch and git safety.
- MUST read `.agents/rules/project-commands.md` before running any project command.

## Rules Reference

Load `.agents/rules/index.md` first, then load only the matching rule files:

- `.agents/rules/debugging.md` for bugs, flaky tests, runtime errors, or unexpected behavior
- `.agents/rules/project-commands.md` before running any project command
- `.agents/rules/context-management.md` for `/compact` decisions, handoff summaries, and provider-specific context control
- `.agents/rules/worktree.md` for branch safety, worktree limits, and ignored path handling

<!-- AGENT_TEAM:END -->

# Claude Instructions

Use this file when working in Claude Code on this repository.

- Keep status updates concise.
- Use `/compact` when the session loses task clarity.
- Run targeted verification before reporting completion.
- Use `agent-team reply-main` when acting as a worker.
