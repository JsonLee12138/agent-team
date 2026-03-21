# Strategic Compact Triggers

## Trigger Contract

Normalize every context-management event to one of these trigger types:

- `manual`: explicit user or controller choice to compact now
- `phase-transition`: the controller is about to leave one logical phase and enter another
- `context-pressure`: the active main thread can no longer clearly hold goal, constraints, decisions, and next step
- `pre-large-read`: the controller is about to read a large diff, large log, large test output, or long worker reply
- `resume-after-pause`: work is resuming after a pause, restart, handoff, or context decay

## Mandatory Triggers

A strategic compact pass is mandatory when any of the following are true:

1. The controller finished one logical phase and is entering the next one.
2. The controller is about to read or paste output that is large enough to displace the current working context.
3. The controller can no longer restate Goal / Phase / Constraints / Next without rereading broad history.
4. The controller is resuming after a pause, handoff, or session restart.

## Workflow-Oriented Mapping

Use `phase-transition` around these controller actions when they meaningfully change what the main session must track:

- `agent-team workflow plan generate`
- `agent-team workflow plan approve`
- `agent-team workflow plan activate`
- `agent-team workflow plan close`

Use `resume-after-pause` when the controller resumes governance planning after a pause or restart.

Use `pre-large-read` before reading:

- large diffs
- long test output
- large logs
- long worker replies
- large workflow/task state dumps

## Worker Trigger Policy

Workers do not treat every trigger as a compact trigger by default.

For workers, strategic compact is only allowed in exception scenarios documented in [main-first.md](main-first.md):

- long-running task that must stay open
- blocked work that needs a durable recovery anchor
- multi-round modification loop where the same worker must continue
