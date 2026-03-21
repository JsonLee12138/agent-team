---
name: tdd
description: "Acceptance-first TDD workflow for solo implementation work. Use when handling a coding change directly and you want to follow a disciplined sequence: read context, define acceptance criteria, write failing tests when feasible, implement, verify, and conclude with passed, failed, or skipped with reason."
---

# TDD

Use this skill as a single-user workflow. Do not introduce worker, task assignment, or team coordination concepts.

## Core Rules

- Read the request, relevant code, and constraints before implementing.
- Define what counts as done before writing production code.
- Prefer a red phase when the task supports automated tests.
- Do not default to E2E. Use E2E only when explicitly requested, already required by the task contract, or when lower-cost verification cannot cover the acceptance risk.
- If the task needs test-case design, regression planning, acceptance coverage, or broader QA execution, invoke `qa-expert` instead of expanding this skill into a full QA workflow.
- If automated tests are not appropriate, define manual verification before implementing.
- Do not treat the work as complete until verification finishes.
- If verification is skipped, state the reason and residual risk explicitly.

## Workflow

Follow this sequence:

1. Read context
2. Define acceptance criteria
3. Write failing tests when feasible
4. Implement
5. Verify
6. Conclude

Do not skip ahead unless the previous step is clearly satisfied.

## Step 1: Read Context

Inspect the request and the surrounding implementation before changing anything.

Confirm:

- goal
- scope
- constraints
- non-goals
- existing verification options

If the request is ambiguous, stop and resolve the ambiguity before continuing.

## Step 2: Define Acceptance Criteria

Write acceptance criteria before implementation.

Cover:

- expected behavior
- edge cases
- failure cases
- observable completion conditions

If acceptance criteria cannot be written yet, the task is not ready for implementation.

## Step 3: Write Failing Tests When Feasible

If the task supports automated verification, write or update tests before implementation.

Target a real red phase:

- new behavior is asserted
- current code fails that assertion
- expected outcome is clear

If automated tests are not a good fit, explicitly define the manual verification procedure before moving on.

## Step 4: Implement

Implement against the acceptance criteria and verification plan.

Do not expand scope during implementation unless the acceptance criteria are updated first.

## Step 5: Verify

Use the strongest available verification path in this order:

1. Existing automated tests
2. New automated tests added for the change
3. Repeatable manual verification with concrete steps and expected results
4. `skipped` only when verification is blocked in the current environment

E2E is not the default verification path. Prefer lower-cost checks first unless E2E is explicitly required or the acceptance risk cannot be covered otherwise.

If the task needs dedicated test-case design, regression execution, or broader QA coverage during verification, invoke `qa-expert`.

If verification fails, return to implementation and fix the issue before concluding.

## Step 6: Conclude

End with exactly one of these outcomes:

- `passed`
- `failed`
- `skipped`

When concluding, include:

- brief summary of what changed
- verification used
- result
- remaining risk if any

If the result is `skipped`, include the reason and the specific risk introduced by not verifying.

## Failure Conditions

Stop and resolve the issue before implementation when:

- requirements are unclear
- acceptance criteria are missing
- no verification strategy is defined

Loop back and keep working when:

- tests fail unexpectedly
- the verification command fails
- manual verification does not satisfy acceptance criteria

## Optional Artifacts

Use lightweight artifacts only when they help:

- `proposal.md`
- `design.md`
- `tests.md`
- explicit verify command

Do not require a fixed directory structure unless the surrounding project already has one.
