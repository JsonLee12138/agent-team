# Debugging Rules

## Trigger

Apply this rule for any bug, flaky test, runtime error, build failure, or unexpected behavior.

## Required Flow

MUST follow the `systematic-debugging` workflow in order. ALWAYS reproduce, inspect, isolate, test, then validate.

### 1. Reproduce First

- MUST capture the exact command, input, environment, and full error text before changing code.
- MUST retry intermittent failures at least 3 times to confirm the pattern.

### 2. Check Logs and Evidence

- ALWAYS read the full stack trace, build output, and related logs before forming a hypothesis.
- ALWAYS inspect recent relevant changes with `git diff` and `git log` when regression is possible.

### 3. Isolate the Cause

- MUST reduce the issue to the smallest reproducible case.
- MUST change one variable at a time when testing a hypothesis.

### 4. Validate the Fix

- MUST rerun the original reproduction steps after the fix.
- MUST run the targeted verification commands for the affected scope.
