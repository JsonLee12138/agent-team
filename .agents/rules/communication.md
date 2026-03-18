# Communication Rules

## Trigger

Apply this rule for all worker-to-controller updates, blockers, handoffs, and completion messages.

## `reply-main` Format

- MUST use `agent-team reply-main "Task completed: <summary>; change archived: <change-name>"` after a successful archive.
- MUST use `agent-team reply-main "Task completed: <summary>; archive failed for <change-name>: <error>"` if archive fails.
- MUST use `agent-team reply-main "Need decision: <problem or options>"` for blockers, ambiguity, or conflicting requirements.
- ALWAYS keep messages factual, single-purpose, and short enough to scan quickly.

## Escalation Protocol

- MUST report blockers immediately when progress depends on a user or controller decision.
- MUST include the concrete problem, what was already checked, and the smallest viable options.
- NEVER hide failed verification, skipped checks, or archive errors.

## Progress Frequency

- MUST send a progress update after each meaningful milestone in long tasks.
- MUST send a progress update at least every 30 minutes when work continues without a milestone.
- MUST send a final completion update only after the task protocol has reached the archive step.
