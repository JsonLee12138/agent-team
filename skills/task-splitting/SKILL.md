---
name: task-splitting
description: "Decompose requirement, design, or brainstorming documents into reviewable task drafts with one function per task and one file per task, then create `agent-team` task packages only after explicit approval. Use when turning a document into tasks, splitting oversized tasks, validating task boundaries, or preparing approved task files for `agent-team task create`."
---

# Task Splitting

Turn source documents into explicit task boundaries. Keep decomposition separate from creation: draft and validate first, create `agent-team` tasks only after explicit user approval.

## Workflow

1. Intake the source document.
2. Extract goals, constraints, deliverables, non-goals, and ambiguities.
3. Decide whether `brainstorming` is needed.
4. Draft tasks by function boundary.
5. Validate every task boundary.
6. Write one task file per task.
7. Ask for confirmation.
8. Create `agent-team` task packages from the approved files.

## 1. Intake the source

- Accept either raw document content or a local document path.
- Read only the files needed for decomposition.
- If the document is missing the goal, deliverable, or target output location for task files, ask before drafting.
- If the user did not specify where task files should go, ask once. If they do not care, place them in a sibling directory next to the source document named `<document-stem>-tasks/`.
- Treat the source document plus user clarifications as the only authority for task scope.

## 2. Decide whether to use `brainstorming`

Invoke `brainstorming` before drafting tasks when any of these are true:

- the document goal is ambiguous
- scope conflicts exist
- several decomposition strategies are equally reasonable
- multiple functions are tangled together
- the user wants discussion before task drafting

Do not force `brainstorming` for a clear, well-bounded document.

## 3. Decompose by function boundary

Apply these rules strictly:

- Prefer one user-visible function or one clear sub-function per task.
- Keep exactly one primary responsibility per task.
- Keep task boundaries explicit.
- Prefer one function per task, not one paragraph per task.
- Reject grouping by document section when that creates mixed scope.
- Do not merge multiple functions just because the document narrates them together.

Split a task when any of these are true:

- it contains multiple peer functions
- it contains more than one acceptance target
- it mixes build and validation responsibilities
- its dependency chain makes independent execution unclear

## 4. Validate every task boundary

A valid task must include all of the following:

- exactly one function or sub-function
- exactly one primary responsibility
- a clear completion condition
- an explicit in-scope section
- an explicit out-of-scope section
- explicit acceptance criteria
- manageable dependencies

If a task fails any check, split or rewrite it before writing files.

## 5. Write one task file per task

- One task must map to one task definition file.
- Treat the task file as the authoritative boundary definition.
- Keep naming stable and reviewable, for example `01-auth-login.md`, `02-auth-session-check.md`.
- Use the template in [references/task-file-template.md](references/task-file-template.md).
- Do not create a combined summary file in place of per-task files.

## 6. Request confirmation before creation

- Present the drafted task set for review.
- Surface ambiguities and unresolved questions.
- Ask for explicit confirmation before creating any `agent-team` tasks.
- If the user asks for changes, update the task files first. Do not redefine scope during the creation phase.

## 7. Create `agent-team` tasks from approved files

When the project exposes the `agent-team` CLI, create one task package per approved task file.

Confirmed command surface:

```bash
agent-team task create --role <role> "<title>" --design "<task-file>"
```

Rules:

- Do not create tasks before approval.
- Preserve title, scope, dependencies, and acceptance exactly from the task file.
- Do not reinterpret or expand scope during creation.
- `--role` is required. Use the role named in the document or task file; otherwise ask the user instead of guessing.
- Report exact success and failure results.
- If creation partially fails, stop and return the created tasks plus the failures explicitly.

## Output contract

### Draft phase output

Return:

- a short summary of the decomposition
- the list of drafted task titles
- the file path for each task file
- any unresolved boundary questions
- an explicit approval request

### Creation phase output

Return:

- each created task title
- its source task file
- the created task identifier if available
- any failure with the exact file that failed

## Guardrails

- Do not implement code changes.
- Do not let the creation phase rewrite scope.
- Do not merge analysis, implementation, and testing into one task.
- Do not force `brainstorming` when the boundaries are already clear.
