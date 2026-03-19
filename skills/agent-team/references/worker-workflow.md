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
   agent-team task done <worker-id> <change-name> <task-id>
   ```
3. Keep commits task-scoped and ready for archive after verification

## Step 4 — Verify

When all tasks are marked done:

```bash
agent-team task verify <worker-id> <change-name>
```

This runs the verify command from `change.yaml` (or `.tasks/config.yaml` default).

## Step 5 — Archive and notify main

After verify completes, follow the required completion chain:

```bash
agent-team task archive <worker-id> <change-name>

# If archive succeeds:
agent-team reply-main "Task completed: <summary>; change archived: <change-name>"

# If archive fails:
agent-team reply-main "Task completed: <summary>; archive failed for <change-name>: <error>"
```

## Rules

- NEVER notify before running verify and attempting archive
- If verify fails, report the failure explicitly and do not claim completion
- If archive fails, still notify main with the archive failure
- If blocked, report immediately: `agent-team reply-main "Need decision: <options>"`
- After completion, any new work must go through a new assign cycle
