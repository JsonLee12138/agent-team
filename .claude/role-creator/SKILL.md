---
name: role-creator
description: >
  Create or update role-specific skill packages with deterministic files.
  Supports output to skills/ (open-source publishing) or agents/teams/ (team use).
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
- `agents/teams/<role-name>/` — for team use with agent-team

## Required Workflow

1. Normalize role input:
   - `role-name` must be kebab-case (default: `frontend-dev` when not provided).
2. Auto-generate role fields first (do not ask user to draft them by default):
   - `description`
   - `system goal`
   - `in-scope` (comma-separated output)
   - `out-of-scope` (comma-separated output)
3. If user rejects the auto result, or AI confidence is low, run the full brainstorming process below.
4. After fields are approved, run skills selection flow.
5. Ask user for target directory (`skills` or `agents/teams`).
6. If target is `agents/teams` and the role already exists in global `~/.claude/skills/`, ask if user wants to copy from there instead of generating.
7. Execute generator script to write files deterministically.

## Brainstorming Trigger Policy

- Default flow: auto-generate role fields first.
- Enter full brainstorming when either condition is met:
  - User says the auto-generated role fields are not acceptable.
  - AI detects low confidence (ambiguous scope, conflicting boundaries, or vague goals).

## Brainstorming Process

1. **Explore context** — scan existing role skills under `skills/*/` and `agents/teams/*/`, templates, and recent commits for patterns.
2. **Ask clarifying questions** — one question per message; prefer multiple choice; focus on role purpose and scope boundaries.
3. **Confirm fields** — present the final `description`, `system_goal`, `in_scope`, `out_of_scope` for user approval before generation.

## Skills Selection UX (After Fields Are Approved)

1. Run `find-skills` first to get recommendation candidates.
2. Show recommendation list in checkbox format (primary) with numbered indices:

   ```text
   Recommended skills:
   1. [ ] skill-a
   2. [ ] skill-b

   Reply in either format:
   - Edit checkboxes to [x]/[ ]
   - Or send numbers, e.g. 1,2
   ```

3. Parse precedence:
   - Checkbox parsing is primary.
   - Numeric parsing (`1,3,5`) is fallback.
   - If both checkbox and numeric forms are present, checkbox result wins.
4. Ask for manual additions after selection:

   ```text
   Any additional skills to add manually? (comma-separated, optional)
   ```

5. Merge selected skills and manual additions with de-duplication, then confirm final list.
6. Final skills may be empty and should be persisted as `skills: []` in `references/role.yaml`.

If `find-skills` is unavailable or returns empty, skip recommendations and ask for manual additions only.

## Generate Command

For open-source publishing (default):

```bash
python3 skills/role-creator/scripts/create_role_skill.py \
  --repo-root . \
  --role-name frontend-dev \
  --description "Frontend role for UI implementation" \
  --system-goal "Ship accessible and maintainable UI work" \
  --in-scope "Build components,Improve accessibility" \
  --out-of-scope "Database migrations,Backend API ownership" \
  --skills "ui-ux-pro-max,better-icons"
```

For team use (agent-team integration):

```bash
python3 skills/role-creator/scripts/create_role_skill.py \
  --repo-root . \
  --role-name frontend-dev \
  --target-dir agents/teams \
  --description "Frontend role for UI implementation" \
  --system-goal "Ship accessible and maintainable UI work" \
  --in-scope "Build components,Improve accessibility" \
  --out-of-scope "Database migrations,Backend API ownership" \
  --skills "ui-ux-pro-max,better-icons"
```

Empty skills example:

```bash
python3 skills/role-creator/scripts/create_role_skill.py \
  --repo-root . \
  --role-name product-manager \
  --description "Product role for roadmap and PRD work" \
  --system-goal "Define clear product requirements and priorities"
```

## Runtime Skill Discovery (Role Usage Guideline)

Generated roles should follow this behavior at runtime:

1. When a role receives a task that requires a skill not listed in its `references/role.yaml`, it should invoke `find-skills` to search for a matching skill.
2. If a suitable skill is found, install it and use it to complete the task.
3. After successful use, suggest adding the skill to the role's `references/role.yaml` for future sessions.
4. If `find-skills` is unavailable or returns no match, the role should notify the user and proceed with best-effort execution.

This ensures roles are self-sufficient and can dynamically extend their capabilities without manual reconfiguration.

## Overwrite Behavior

- If the target directory exists, the script asks for confirmation before overwrite.
- On confirmation, it creates a backup at:
  - `<parent>/.backup/<role-name>-<timestamp>/`
- Then it overwrites managed files only (`SKILL.md`, `references/role.yaml`, `system.md`).

## Validation

Run after changes:

```bash
python3 -m unittest skills/role-creator/tests/test_create_role_skill.py -v
```
