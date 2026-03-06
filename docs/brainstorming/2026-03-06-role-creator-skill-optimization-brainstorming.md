# Brainstorming: role-creator SKILL.md Optimization

**Role**: General Strategist
**Date**: 2026-03-06

## Problem Statement

`skills/role-creator/SKILL.md` has accumulated several issues:
1. Documentation-implementation mismatch (role.yaml field names, CLI flags)
2. Over-engineered sections duplicating Go code internals (checkbox/numeric parsing, 4-directory scanning)
3. Built-in brainstorming process that duplicates the standalone `/brainstorming` skill
4. Incomplete CLI flag documentation

## Goals

- Fix all inaccuracies between SKILL.md and CLI implementation
- Simplify by removing implementation-detail descriptions that belong in Go code
- Restructure around the actual execution workflow (chapter order = execution order)
- Maintain single-file simplicity

## Constraints & Assumptions

- Output remains a single SKILL.md file (no file split)
- The Go CLI (`cmd/role_create.go`, `internal/role_create.go`) and templates are the source of truth
- Existing roles in `.agents/teams/` should remain compatible

## Candidate Approaches

### Approach A: Incremental Fix
Fix inaccuracies in-place, delete bloated sections, keep same structure.
- **Pro**: Minimal diff, easy to review
- **Con**: Chapter ordering stays suboptimal (brainstorming policy in the middle, Generate Command at the end)

### Approach B: Restructure by Workflow (Recommended)
Reorganize the entire file around a 5-step workflow that matches execution order.
- **Pro**: Chapter order = execution order; eliminates all issues naturally; AI reads top-to-bottom and executes
- **Con**: Full rewrite, larger diff

### Approach C: Two-file Split
Split into SKILL.md (AI behavior) + references/cli-reference.md (CLI details).
- **Pro**: Best separation of concerns
- **Con**: Adds maintenance overhead for a single-skill directory; AI needs extra file read

## Recommended Design

**Approach B** — Restructure by Workflow.

### New Document Structure

```
Frontmatter (name, description, triggers)
+-- # role-creator (intro + managed files + target dirs)
+-- ## Required Workflow (5-step overview)
+-- ## Step 1: Normalize Input (kebab-case validation)
+-- ## Step 2: Generate Role Fields (auto-generate + /brainstorming delegation)
+-- ## Step 3: Select Skills (find-skills -> user selection -> manual additions)
+-- ## Step 4: Execute CLI
|   +-- ### CLI Reference (full flag table)
|   +-- ### Examples (common scenarios)
+-- ## Step 5: Validate Output (actual template structure)
+-- ## Overwrite Behavior (backup logic, --overwrite flag)
+-- ## Runtime Skill Discovery (runtime guideline, preserved)
```

### Key Decisions

| Topic | Decision | Rationale |
|---|---|---|
| Brainstorming | Delegate to `/brainstorming` | Avoids duplication with standalone skill |
| Skills Selection UX | Describe AI interaction only, not parsing details | Go code handles checkbox/numeric parsing |
| Local Skill Resolution | Remove | CLI handles directory scanning internally |
| CLI flags | Document all 14 parameters in a table | User requested complete listing |
| role.yaml validation | Match actual Go template output | Fix `system_goal` -> `system_prompt_file`, flat -> nested `scope.*` |
| Overwrite Behavior | Preserve, link to `--overwrite` flag | Still accurate and useful |
| Runtime Skill Discovery | Preserve unchanged | Role usage guideline, not generation logic |

### Section Details

#### Frontmatter + Intro
- Keep existing trigger keywords
- Simplify intro: "three managed files" + target directory options

#### Step 1: Normalize Input
- kebab-case validation rule
- Default suggestion on invalid input

#### Step 2: Generate Role Fields
- Auto-generate `description`, `system goal`, `in-scope`, `out-of-scope`
- Present to user for approval
- Rejection -> delegate to `/brainstorming`
- After approval, ask for target directory (`skills` or `.agents/teams`)

#### Step 3: Select Skills
- `find-skills` recommendations -> user selection -> manual additions -> confirm
- Empty skills list is valid
- Graceful fallback if `find-skills` unavailable

#### Step 4: Execute CLI — CLI Reference

| Flag | Type | Required | Default | Description |
|---|---|---|---|---|
| `<role-name>` | arg | yes | — | Role name (kebab-case) |
| `--description` | string | yes | — | Role description |
| `--system-goal` | string | yes | — | Primary objective for system.md |
| `--in-scope` | string[] | no | (from description) | In-scope items (repeatable, comma-separated) |
| `--out-of-scope` | string[] | no | fallback text | Out-of-scope items (repeatable, comma-separated) |
| `--skills` | string | no | `""` | Final selected skills (comma-separated) |
| `--recommended-skills` | string | no | `""` | Recommended skills from find-skills |
| `--add-skills` | string | no | `""` | Skills to add on top of selection |
| `--remove-skills` | string | no | `""` | Skills to remove from candidate list |
| `--manual-skills` | string | no | `""` | Manual fallback when recommendations unavailable |
| `--target-dir` | string | no | `skills` | Target: `skills`, `.agents/teams`, or custom |
| `--overwrite` | string | no | `ask` | Overwrite mode: `ask`/`yes`/`no` |
| `--repo-root` | string | no | `.` | Repository root path |
| `--force` | bool | no | `false` | Skip global duplicate check |

Skills resolution priority: `--skills` > `--recommended-skills` > `--manual-skills`, then `--add-skills` appended, `--remove-skills` excluded.

#### Step 5: Validate Output
Verify three files against actual template structure:
- `SKILL.md` — role name + description + triggers
- `references/role.yaml` — `name`, `description`, `system_prompt_file: system.md`, `scope.in_scope`, `scope.out_of_scope`, `constraints.single_role_focus: true`, `skills`
- `system.md` — system goal + operating constraints

#### Overwrite Behavior
- Controlled by `--overwrite` flag (ask/yes/no)
- Backup at `<parent>/.backup/<role-name>-<timestamp>/`
- Only managed files are overwritten

#### Runtime Skill Discovery
- Role invokes `find-skills` at runtime for missing skills
- Ask user for global vs project-level install
- Default to global

## Risks & Mitigations

| Risk | Mitigation |
|---|---|
| Full rewrite may introduce new errors | Diff against existing roles to verify correctness |
| Delegating to `/brainstorming` adds skill dependency | `/brainstorming` is a core skill always present |
| Removing parsing details may confuse contributors | Go code is well-documented with tests |

## Validation Strategy

1. Diff new SKILL.md against current to confirm all removals are intentional
2. Verify CLI flag table matches `cmd/role_create.go` flag definitions exactly
3. Verify role.yaml field list matches `internal/templates/role.yaml.tmpl`
4. Test: create a new role with the documented examples, inspect output

## Open Questions

None — all decisions confirmed through brainstorming dialogue.
