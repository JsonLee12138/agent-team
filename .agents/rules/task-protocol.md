# Task Protocol Rules

## Trigger

Apply this rule when a change is assigned, implemented, verified, completed, or handed back to the controller.

## Required Completion Chain

- MUST finish implementation and run the required verification before preparing the final handoff.
- MUST review `git status` and stage only task-scoped files.
- MUST commit task-scoped changes before archive when uncommitted task work exists.
- MUST run `agent-team task archive <worker-id> <change-name>` after the commit step.
- MUST run `agent-team reply-main` after every archive attempt, including failure cases.
- MUST NOT start another task before the completion message has been sent.

## Failure Handling

- MUST report verify failures explicitly and include the failing command or reason.
- MUST report archive failures explicitly and still notify main with the failure details.
- NEVER claim completion while the change is still uncommitted or unreported.
