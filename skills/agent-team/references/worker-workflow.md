# Worker TDD Workflow

When a `[New Change Assigned]` notification arrives, follow this cycle:

## Step 1 — Understand the task

1. Read `.tasks/changes/<change-name>/proposal.md` (requirements)
2. Read `.tasks/changes/<change-name>/design.md` if present (architecture decisions)
3. Read `change.yaml` to see the task list and verify config

## Step 2 — Write acceptance tests first (TDD)

Before implementing, define what "done" means:

1. Create or update `.tasks/changes/<change-name>/tests.md` with acceptance criteria
2. Write failing tests in code if verify command is configured (e.g., `go test ./...`)
   - Tests MUST fail at this point (red phase)

## Step 3 — Implement

1. Implement each task in `change.yaml`
2. Mark each task done as you complete it:
   ```bash
   agent-team task done <change-name> <task-id>
   ```
3. Keep committing regularly with clear messages

## Step 4 — Verify

When all tasks are marked done:

```bash
agent-team task verify <change-name>
```

This runs the verify command from `change.yaml` (or `.tasks/config.yaml` default).

## Step 5 — Notify main

After verify completes:

```bash
# If verify passed:
agent-team reply-main "Task completed: <summary>; verify: passed"

# If verify failed (describe what failed):
agent-team reply-main "Task completed: <summary>; verify: failed — <reason>"

# If verify is skipped (no verify config):
agent-team reply-main "Task completed: <summary>; verify: skipped"
```

## Rules

- NEVER notify before running verify (or explicitly skipping it)
- If verify fails, fix and re-run before notifying
- If blocked, report immediately: `agent-team reply-main "Need decision: <options>"`
- After completion, any new work must go through a new assign cycle
