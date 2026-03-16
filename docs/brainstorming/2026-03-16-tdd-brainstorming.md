# TDD Skill Brainstorming

## Role

general strategist

## Problem Statement And Goals

Current TDD guidance exists inside `skills/agent-team/`, but it is embedded in the task and worker workflow. The goal is to extract that working method into a standalone `tdd` skill for direct use by the controller or a single operator. This new skill should preserve the existing TDD shape of the workflow while removing all worker, task assignment, and team-collaboration concepts.

Goals:

- Create a standalone `tdd` skill with a simple, reusable single-user workflow
- Keep the workflow aligned with the current pattern: read requirements, define acceptance, test first when possible, implement, verify, conclude
- Avoid any required changes to `agent-team` or its existing task lifecycle
- Keep the skill usable for normal coding work, including tasks that cannot be fully automated

## Constraints And Assumptions

- `agent-team` remains unchanged
- The new skill is manually invoked and not wired into `agent-team`
- The skill is intended for direct use by the main controller or a solo operator
- The skill should still deserve the name `tdd`, so it cannot degrade into a generic checklist
- Some task types support automated tests; others only support manual verification

## Candidate Approaches With Trade-Offs

### Approach 1: Light Process TDD

Flow: read context -> write acceptance criteria -> write failing tests when feasible -> implement -> verify -> conclude.

Pros:

- Closest to the current workflow already used in `agent-team`
- Flexible enough for real-world tasks that do not always fit strict red-green-refactor
- Keeps TDD intent without making the skill brittle

Cons:

- Requires judgment on when automated tests are feasible
- Slightly less strict than pure TDD

### Approach 2: Strict Red-Green TDD

Flow: no implementation is allowed until automated failing tests exist.

Pros:

- Strongest TDD discipline
- Very clear enforcement model

Cons:

- Poor fit for docs, scripts, ops, environment, and exploratory tasks
- Would make the standalone skill too narrow for normal controller use

### Approach 3: Acceptance-First Verification

Flow: define acceptance and verification first, but failing tests are optional rather than strongly expected.

Pros:

- Broadest applicability
- Easy to adopt

Cons:

- Too soft for a skill named `tdd`
- Risks drifting into a generic “plan then implement” workflow

## Recommended Design

Use Approach 1: Light Process TDD.

This keeps the existing workflow logic intact while removing team-specific concepts. The skill should define a disciplined sequence:

1. Read the request, related code, and relevant constraints
2. Define acceptance criteria before implementation
3. If the work supports automation, write failing tests first
4. Implement against the acceptance criteria
5. Run verification
6. Conclude with `passed`, `failed`, or `skipped` with reason

Core principles:

- Define “done” before coding
- Prefer red phase when technically reasonable
- If automation is not appropriate, define explicit manual verification before implementation
- Verification failure means the work is not done
- Any skipped verification must include reason and residual risk

## Architecture

The skill should be organized as a single-user workflow with the following stages:

- `Context Read`: gather requirements, code context, constraints, and non-goals
- `Acceptance First`: write completion criteria, edge cases, and failure conditions
- `Test First When Possible`: create failing tests when the task supports automation
- `Implement`: perform the change only after acceptance and verification strategy are clear
- `Verify`: run tests or another explicit validation procedure
- `Conclude`: summarize outcome as passed, failed, or skipped with reason

This architecture intentionally excludes:

- Worker lifecycle concepts
- Task assignment semantics
- Team communication protocols
- Merge or review automation

## Components

The skill needs only a minimal set of responsibilities:

- Intake and context reading
- Acceptance criteria shaping
- Verification strategy definition
- TDD-first implementation discipline
- Result conclusion and reporting

These are conceptual responsibilities inside the skill instructions, not separate code modules.

## Data Flow

The expected single-user data flow is:

1. Receive a task or change request
2. Read surrounding code and constraints
3. Produce acceptance criteria
4. Produce failing automated tests when feasible, otherwise produce manual verification steps
5. Implement the change
6. Run verification
7. Report result and remaining risk

Optional artifacts the skill may encourage, but should not hard-require:

- `proposal.md`
- `design.md`
- `tests.md`
- a verify command or explicit manual validation checklist

## Error Handling

The skill should explicitly stop progress in these cases:

- Requirements are unclear
- Acceptance criteria cannot yet be written
- Verification strategy is still undefined

The skill should explicitly loop back in these cases:

- Automated tests fail after implementation
- Verification command fails
- Manual verification reveals unmet acceptance criteria

The skill should explicitly allow `skipped` only when:

- Verification cannot be executed in the current environment
- The reason is stated clearly
- Remaining risk is surfaced clearly

## Validation And Test Strategy

Verification priority should be:

1. Automated tests already present in the project
2. New automated tests added for the requested change
3. Repeatable manual verification with concrete steps and expected outcomes
4. `skipped` with documented reason and risk only when verification is blocked

For tasks that support automated tests, the intended flow is red -> implement -> green.

For tasks that do not support automated tests well, the intended flow is acceptance-first -> implement -> manual verify.

## Risks And Mitigations

- Risk: the skill becomes too generic and loses TDD identity
  - Mitigation: keep “acceptance first” and “test first when possible” as mandatory sequence rules

- Risk: users treat verify as optional
  - Mitigation: explicitly forbid completion without `passed`, `failed`, or `skipped with reason`

- Risk: strict TDD expectations make the skill unusable for some tasks
  - Mitigation: allow manual verification as a first-class fallback, but require it to be explicit before implementation

## Open Questions

- None at this stage
