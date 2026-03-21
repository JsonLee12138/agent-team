---
name: role-creator
description: >
  Create or update role-specific skill packages with deterministic files.
  Supports output to skills/ (open-source publishing) or .agent-team/teams/ (team use).
  Triggers: 创建角色, 新建 role, create role, 更新 role scope, edit role, update role,
  add role skill, 修改角色配置.
  Use when the user asks to create, update, or edit frontend/backend/product (or custom) role
  skills with auto-generated role fields, guided brainstorming fallback, and curated skills selection.
---

# role-creator

Generate a role skill package in a fixed contract:
- `<target>/SKILL.md`
- `<target>/references/role.yaml`
- `<target>/system.md`

Target directory options:
- `skills/<role-name>/` — for open-source publishing (default)
- `.agent-team/teams/<role-name>/` — for team use with agent-team
- `<custom-path>/<role-name>/` — for custom local output paths

## Required Workflow

1. **Normalize Input** — validate role name as kebab-case.
2. **Generate Role Fields** — auto-generate fields, delegate to `/brainstorming` on rejection.
3. **Select Skills** — `find-skills` recommendations, user selection, manual additions.
4. **Execute CLI** — run `agent-team role create` with approved parameters.
5. **Validate Output** — verify three managed files match expected template structure.

## Step 1: Normalize Input

- `role-name` must be kebab-case. The current CLI requires it as a positional argument.
- If the input is not kebab-case, suggest a normalized version and confirm with user.

## Step 2: Generate Role Fields

1. Auto-generate the following fields (do not ask user to draft them by default):
   - `description`
   - `system goal`
   - `in-scope` (comma-separated)
   - `out-of-scope` (comma-separated)
2. Present generated fields for user approval.
3. If user rejects, or AI confidence is low (ambiguous scope, conflicting boundaries, vague goals), delegate to `/brainstorming` skill for structured refinement.
4. After fields are approved, ask user for target directory (`skills`, `.agent-team/teams`, or a custom path).
5. If target is `skills` or `.agent-team/teams`, note that the CLI checks matching global roles in `~/.agents/roles/` unless `--force` is passed. If matches are found, the CLI prompts whether to continue creating the local role.

## Step 3: Select Skills

1. Run `find-skills` to get recommendation candidates.
2. Show recommendation list and let user select desired skills.
3. Ask for manual additions after selection.
4. When confirming the final skill identifiers, prefer the full remote identifier if the skill is known from a remote source (for example `jsonlee12138/prompts@design-patterns-principles`).
5. Only use a local short name (for example `design-patterns-principles`) when no matching remote skill is found and the skill exists only locally.
6. Merge selected + manual additions with de-duplication, then confirm final list.
7. Final skills may be empty — will be persisted as `skills: []` in `references/role.yaml`; otherwise persist `skills[{name, description}]` using the preferred identifier rule above.

If `find-skills` is unavailable or returns empty, skip recommendations and ask for manual additions only. In that fallback path, still prefer a full remote identifier if the user already provides one; otherwise use a local short name only for local-only skills.

## Step 4: Execute CLI

### CLI Reference

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
| `--target-dir` | string | no | `skills` | Target: `skills`, `.agent-team/teams`, or custom path |
| `--overwrite` | string | no | `ask` | Overwrite mode: `ask`/`yes`/`no` |
| `--repo-root` | string | no | `.` | Repository root path |
| `--force` | bool | no | `false` | Skip global duplicate check |

Skills resolution priority: `--skills` > `--recommended-skills` > `--manual-skills`, then `--add-skills` appended, `--remove-skills` excluded.

### Examples

For open-source publishing (default):

```bash
agent-team role create frontend-dev \
  --description "Frontend role for UI implementation" \
  --system-goal "Ship accessible and maintainable UI work" \
  --in-scope "Build components,Improve accessibility" \
  --out-of-scope "Database migrations,Backend API ownership" \
  --skills "ui-ux-pro-max,better-icons"
```

For team use (agent-team integration):

```bash
agent-team role create frontend-dev \
  --target-dir .agent-team/teams \
  --description "Frontend role for UI implementation" \
  --system-goal "Ship accessible and maintainable UI work" \
  --in-scope "Build components,Improve accessibility" \
  --out-of-scope "Database migrations,Backend API ownership" \
  --skills "ui-ux-pro-max,better-icons"
```

Empty skills example:

```bash
agent-team role create product-manager \
  --description "Product role for roadmap and PRD work" \
  --system-goal "Define clear product requirements and priorities"
```

## Step 5: Validate Output

Verify three files against actual template structure:

1. **`SKILL.md`** — contains the role name, generated description frontmatter, and links to `references/role.yaml` and `system.md`.
2. **`references/role.yaml`** — must contain:
   - `name` — role name
   - `description` — role description
   - `system_prompt_file: system.md`
   - `scope.in_scope` — list of in-scope items
   - `scope.out_of_scope` — list of out-of-scope items
   - `constraints.single_role_focus: true`
   - `skills` — selected skills as objects with `name` and `description` (or empty `[]`)
3. **`system.md`** — contains system goal and operating constraints.

If any file is missing or contains unexpected content relative to the current templates, report the discrepancy and offer to regenerate.

## Overwrite Behavior

- Controlled by `--overwrite` flag (`ask`/`yes`/`no`).
- Current CLI does not create a backup directory when overwriting.
- Only managed files are overwritten (`SKILL.md`, `references/role.yaml`, `system.md`); other files in the role directory are preserved.
- Legacy root-level `role.yaml` is removed if present after generation.

## Runtime Skill Discovery (Role Usage Guideline)

Generated roles should follow this behavior at runtime:

1. When a role receives a task that requires a skill not listed in its `references/role.yaml`, it should invoke `find-skills` to search for a matching skill.
2. If a suitable skill is found, it may only attempt project-level installation or use an already available local/project-cached skill.
3. Global installation is not allowed for worker/runtime resolution.
4. If the skill is still unavailable locally, emit a warning with the reason and suggested next step, then continue with best-effort execution.
5. After successful use, suggest adding the skill to the role's `references/role.yaml` for future sessions by regenerating the role.
6. If `find-skills` is unavailable or returns no match, the role should notify the user and proceed with best-effort execution.
