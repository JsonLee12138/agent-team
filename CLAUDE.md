<!-- AGENT_TEAM:START -->
# Claude Instructions

Use this file when working in Claude Code on this repository.

- MUST read `.agents/rules/index.md` at task start and load only the rule files required by the task.
- MUST follow `.agents/rules/context-management.md` for context-cleanup and index-first recovery whenever context drifts, phases change, or work resumes.
- MUST keep status updates concise.
- MUST obey `.agents/rules/worktree.md` for branch and git safety.
- MUST read `.agents/rules/project-commands.md` before running any project command.
- MUST follow `.agents/rules/agent-team-commands.md` for worker lifecycle and generated-file boundaries.
- MUST follow `.agents/rules/merge-workflow.md` for controller-side rebase and merge sequencing.

## Rules Reference

Load `.agents/rules/index.md` first, then load only the matching rule files:

- `.agents/rules/debugging.md` for bugs, flaky tests, runtime errors, or unexpected behavior
- `.agents/rules/project-commands.md` before running any project command
- `.agents/rules/agent-team-commands.md` for agent-team CLI boundaries and worker lifecycle operations
- `.agents/rules/merge-workflow.md` for controller-side rebase, merge ordering, and generated file safety
- `.agents/rules/context-management.md` for context-cleanup triggers, session reset boundaries, and index-first file recovery
- `.agents/rules/worktree.md` for branch safety, worktree limits, and ignored path handling

<!-- AGENT_TEAM:END -->
