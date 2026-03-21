# Task Splitting Skill Brainstorming

- Date: 2026-03-21
- Role: general strategist
- Save location: `docs/brainstorming/2026-03-21-task-splitting-skill-brainstorming.md`

## Problem Statement

Create a skill for task creation and task decomposition from documents.

The skill must:
- decompose tasks from documents
- follow single-function principle
- follow single-responsibility principle
- keep task boundaries explicit
- prefer one function per task
- enforce one task per file
- support a two-phase flow: draft first, create `agent-team` tasks after confirmation
- invoke `brainstorming` when the document is ambiguous or boundaries are hard to decide

## Goals

1. Turn requirement/design documents into reviewable task drafts.
2. Ensure each task covers exactly one clear function or sub-function.
3. Generate one task definition file per task.
4. After user confirmation, create corresponding `agent-team` tasks.
5. Keep decomposition logic and task creation logic separated.

## Non-Goals

- Do not directly implement code changes.
- Do not merge multiple functions into one task just because they appear in the same document section.
- Do not let the creation phase redefine task scope.
- Do not force `brainstorming` for every document; use it only when needed.

## Constraints and Assumptions

- Boundary priority is by function, not by paragraph or file layout.
- A task file is the authoritative boundary definition.
- `brainstorming` is the escalation path when goals or scope are unclear.
- The skill uses a dual-mode flow:
  - mode 1: produce task drafts / task files
  - mode 2: create `agent-team` tasks after approval
- One task must map to one task definition file.

## Candidate Approaches

### Option A — Two-Phase Decomposition and Creation (Recommended)

Flow:
1. Read document and extract goals, constraints, deliverables, non-goals, ambiguities.
2. Decide whether `brainstorming` is needed.
3. Produce task drafts.
4. Validate boundaries.
5. Generate one task file per task.
6. Ask for confirmation.
7. Create `agent-team` tasks.

Advantages:
- clearest boundaries
- easy review before write actions
- lower risk of bad task creation
- matches the required dual-mode flow

Trade-offs:
- one extra confirmation step

### Option B — Direct Creation from Document

Flow:
1. Read document.
2. Decompose immediately.
3. Create `agent-team` tasks directly.

Advantages:
- fast

Trade-offs:
- high risk of incorrect boundaries
- harder to correct after creation
- weak review checkpoint

### Option C — Structured Spec First, Then Generate Tasks

Flow:
1. Normalize the document into a structured intermediate spec.
2. Generate tasks from the spec.
3. Create `agent-team` tasks.

Advantages:
- very systematic
- good for large, complex projects

Trade-offs:
- heavier workflow
- likely over-designed for current scope

## Recommended Design

Choose **Option A**.

Reason:
- best fit for one-function-per-task
- best fit for one-task-per-file
- keeps review and creation separate
- supports ambiguity handling through `brainstorming`

## Architecture

### 1. Document Intake
- Accept document content or document path.
- Extract goals, constraints, deliverables, non-goals, and ambiguities.
- Stop and ask follow-up questions if required inputs are missing.

### 2. Brainstorm Gate
- Determine whether `brainstorming` is needed.
- Trigger when:
  - goals are ambiguous
  - scope conflicts exist
  - multiple decomposition strategies are equally reasonable
  - several functions are tangled together
  - the user explicitly wants discussion first

### 3. Task Decomposer
- Decompose by function boundary first.
- Prefer one user-visible function or one clear sub-function per task.
- Reject paragraph-based or document-section-based grouping when it creates mixed scope.

### 4. Boundary Validator
- Validate each task for:
  - single function
  - single responsibility
  - explicit in-scope
  - explicit out-of-scope
  - explicit acceptance criteria
  - manageable dependencies
- Split or rewrite any invalid task.

### 5. Task File Generator
- Generate one file per task.
- Proposed file content:
  - title
  - goal
  - in-scope
  - out-of-scope
  - dependencies
  - acceptance

### 6. Agent-Team Task Creator
- After approval, create `agent-team` tasks from task files.
- Preserve title, scope, dependencies, and acceptance exactly.
- Do not re-interpret or expand scope during creation.

## Execution Flow

1. Accept document input.
2. Extract requirement structure.
3. Evaluate whether `brainstorming` is needed.
4. Produce task drafts.
5. Validate task boundaries.
6. Generate task files.
7. Request user confirmation.
8. Create `agent-team` tasks.

## State Model

### source
Original document and user supplements.

### draft
Task drafts, boundary notes, and unresolved questions.

### final
Approved task files and created `agent-team` tasks.

## Boundary Rules

A valid task must satisfy all of the following:
- exactly one function or sub-function
- exactly one primary responsibility
- clear completion condition
- clear out-of-scope statement
- no mixed analysis + implementation + testing responsibility in the same task

A task should be split when:
- it contains multiple peer functions
- it contains more than one acceptance target
- it mixes build and validation responsibilities
- its dependency chain makes independent execution unclear

A task should not be merged only because:
- the items appear in the same paragraph
- the items touch the same broad feature area
- the document author described them together narratively

## Error Handling and Exceptions

- Missing document: stop and ask for document or path.
- Document with goals but no boundaries: ask follow-up questions; use `brainstorming` if still unclear.
- Oversized task: split further until it becomes one function per task.
- Over-fragmented task set: allow regrouping only if the regrouped task still keeps a single functional goal and remains independently acceptable.
- Task file generation failure: stop before creating `agent-team` tasks.
- Partial creation failure: return exact success/failure results; do not silently continue.

## Validation Strategy

### Single Function Check
Ensure each task targets one functional outcome.

### Single Responsibility Check
Ensure each task does not mix analysis, implementation, and testing responsibilities.

### Boundary Completeness Check
Ensure each task file contains:
- goal
- in-scope
- out-of-scope
- dependencies
- acceptance

### Creation Consistency Check
Ensure created `agent-team` tasks match task files exactly.

## Naming Guidance

When referencing the brainstorming capability in this skill, use `brainstorming` directly. Do not use the prefixed form.

## Risks and Mitigations

| Risk | Mitigation |
| --- | --- |
| Document ambiguity causes poor decomposition | Ask follow-up questions or invoke `brainstorming` |
| Multiple functions slip into one task | Enforce Boundary Validator split rules |
| Draft and created task mismatch | Treat task file as single source of truth |
| User confirmation skipped | Require explicit confirmation gate before creation |
| Narrative document structure distorts boundaries | Extract structured goals and non-goals before decomposition |

## Open Questions

None at this stage.

## Final Recommendation

Build the skill around this rule:

> First produce reviewable task drafts and one task file per task from the source document. Only after confirmation should the skill create formal `agent-team` tasks. If boundaries are unclear, enter `brainstorming` first.
