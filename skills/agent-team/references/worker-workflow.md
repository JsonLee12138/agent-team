# Worker Delivery Workflow (Governance-Only)

When a `[New Assignment]` notification arrives, follow this cycle:

## Step 1 — Understand the assignment

1. Read task context provided by controller messages and referenced docs.
2. Clarify blockers early with `agent-team reply-main`.

## Step 2 — Define acceptance first (TDD preferred)

Before implementing, define what "done" means:

1. Capture acceptance criteria in tracked files for this assignment.
2. If feasible, write failing tests first.

## Step 3 — Implement

1. Implement assignment scope only.
2. Keep commits task-scoped.

## Step 4 — Verify

Run the required verification commands for the changed scope (for this repo typically `make test` / targeted tests).

## Step 5 — Notify main

After verification and commit:

```bash
agent-team reply-main "Task completed: <summary>"
```

If blocked or ambiguous:

```bash
agent-team reply-main "Need decision: <problem or options>"
```

## Rules

- NEVER claim completion before verification evidence exists.
- If verify fails, report failure explicitly and do not claim completion.
- Keep messages concise and factual.
- After completion, any new work must go through a new assign cycle.
