# Task File Template

Use one Markdown file per task. Keep the task boundary explicit and reviewable.

## Required structure

```markdown
# <Task title>

- Status: draft
- Role: <required for creation, otherwise TBD>
- Source: <source document path or note>

## Goal

<One functional outcome only>

## In Scope

- <item>
- <item>

## Out of Scope

- <item>
- <item>

## Dependencies

- None

## Acceptance

- [ ] <observable completion condition>
- [ ] <observable completion condition>
```

## Writing rules

- `Goal` must describe one function or one clear sub-function.
- `In Scope` must list only work required for that function.
- `Out of Scope` must state adjacent work that is intentionally excluded.
- `Dependencies` must stay minimal and explicit.
- `Acceptance` must be observable and finite.
- If a task needs two unrelated acceptance targets, split it.
- If a task mixes build and validation responsibilities, split it.

## Naming guidance

Prefer stable, ordered, kebab-case filenames:

- `01-parse-requirements.md`
- `02-validate-boundaries.md`
- `03-create-task-package.md`

## Creation mapping

When creating approved `agent-team` tasks, pass the task file path through `--design` so the created task package points back to the authoritative task boundary document.
