# Communication Rules

## Trigger

Apply this rule for all worker-to-controller updates, blockers, handoffs, and completion messages.

## `reply-main` Format

- MUST use `agent-team reply-main "Task completed: <summary>; change archived: <change-name>"` after a successful archive.
- MUST use `agent-team reply-main "Task completed: <summary>; archive failed for <change-name>: <error>"` if archive fails.
- MUST use `agent-team reply-main "Need decision: <problem or options>"` for blockers or ambiguity.
- ALWAYS keep messages factual, single-purpose, and short enough to scan quickly.

## Escalation Protocol

- MUST report blockers immediately when progress depends on a user or controller decision.
- NEVER hide failed verification, skipped checks, or archive errors.
