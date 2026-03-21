# Integration Contract

## Purpose

Strategic compact is a shared policy skill used by controller-facing workflows.
It should be invoked by rules and controller skills, not embedded as duplicated strategy text in each place.

## Rule Layer

`.agents/rules/context-management.md` should answer only:

1. when context management becomes mandatory
2. which skill must be used

It should not carry provider-specific strategy branches or a full compact algorithm.

## Workflow Skill Integration

`skills/workflow/SKILL.md` should invoke or reference strategic compact at these points:

- before/after a meaningful phase transition driven by `workflow plan generate|approve|activate|close`
- before the controller reads large workflow output, logs, test results, or diffs
- when resuming an existing run-state after pause or restart

Workflow v1 integration is documentation-first.
It does not require `cmd/workflow.go` or `internal/workflow.go` to auto-trigger compact.

## Agent-Team Skill Integration

`skills/agent-team/SKILL.md` should align controller behavior with these rules:

- main evaluates `phase-transition` after worker `reply-main` if review or re-dispatch will follow
- main evaluates `pre-large-read` before reading a large diff, log, or long worker reply
- main evaluates `resume-after-pause` when returning to an existing worker/controller thread after context decay
- worker standard completion does not route through compact by default

## Command Primitive

Strategic compact reuses the existing command surface:

```bash
agent-team compact --to main
```

Optional annotation is allowed:

```bash
agent-team compact --to main --message "Goal: ... Next: ..."
```

Do not create a new pane-resolution contract here. Pane targeting belongs to the existing compact command.

## Future Work Boundary

Possible future automation hooks may exist around:

- workflow node transitions
- task lifecycle transitions

But those hooks are not part of v1. v1 only establishes the policy and invocation contract.
