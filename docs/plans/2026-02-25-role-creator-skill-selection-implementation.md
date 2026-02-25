# Role Creator Skill Selection UX Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Update `role-creator` so skill selection is checkbox-first with numeric fallback, supports manual additions, and allows empty final skills persisted as `skills: []`.

**Architecture:** Keep behavior changes tightly scoped to the role-creator package. Update instruction text in `SKILL.md` for deterministic user interaction, then make a minimal generator change so an empty final list is valid and serialized correctly in `role.yaml`. Use tests first for behavior changes and keep template/render changes localized.

**Tech Stack:** Python 3 (`unittest`, `argparse`, `string.Template`), Markdown templates, OpenSpec validation.

---

### Task 1: Allow empty final skills in generator output

**Files:**
- Modify: `skills/role-creator/tests/test_create_role_skill.py`
- Modify: `skills/role-creator/scripts/create_role_skill.py`
- Modify: `skills/role-creator/assets/templates/role.yaml.tmpl`

**Step 1: Write failing tests**

Add tests to `skills/role-creator/tests/test_create_role_skill.py`:

```python
def test_empty_skills_allowed(self):
    final = self.module.resolve_final_skills(
        selected_skills=[],
        recommended_skills=[],
        added_skills=[],
        removed_skills=[],
        manual_skills=[],
    )
    self.assertEqual(final, [])

def test_role_yaml_uses_empty_array_for_empty_skills(self):
    config = self.module.RoleConfig(
        role_name="frontend-dev",
        description="Frontend role",
        system_goal="Ship UI",
        in_scope=["Build components"],
        out_of_scope=["Own backend"],
        skills=[],
    )
    rendered = self.module.render_files(config, self.templates_dir)
    self.assertIn("skills: []", rendered["role.yaml"])
```

**Step 2: Run tests to verify they fail**

Run:

```bash
python3 -m unittest skills/role-creator/tests/test_create_role_skill.py -v
```

Expected:
- At least one failure showing empty-skills behavior is currently rejected or serialized incorrectly.

**Step 3: Write minimal implementation**

In `skills/role-creator/scripts/create_role_skill.py`:
- Update `resolve_final_skills(...)` to allow returning `[]` when no selected/recommended/manual skills exist.
- Keep add/remove handling deterministic.
- Add a rendered substitution field for role YAML skills, for example `skills_field`, where:
  - non-empty skills => multiline list under `skills:`
  - empty skills => exact `skills: []`

In `skills/role-creator/assets/templates/role.yaml.tmpl`:
- Replace the current `skills` block placeholder with a single `${skills_field}` placeholder.

Suggested rendering shape:

```python
def render_skills_field(skills: list[str]) -> str:
    if not skills:
        return "skills: []"
    return "skills:\n" + "\n".join(f'  - {yaml_quote(s)}' for s in skills)
```

**Step 4: Run tests to verify they pass**

Run:

```bash
python3 -m unittest skills/role-creator/tests/test_create_role_skill.py -v
```

Expected:
- All tests pass.

**Step 5: Commit**

```bash
git add skills/role-creator/tests/test_create_role_skill.py skills/role-creator/scripts/create_role_skill.py skills/role-creator/assets/templates/role.yaml.tmpl
git commit -m "feat: allow empty role skills and serialize as empty array"
```

### Task 2: Update SKILL.md to checkbox-first selection UX

**Files:**
- Modify: `skills/role-creator/SKILL.md`

**Step 1: Write a failing content check**

Before editing, verify required phrases do not yet exist:

```bash
rg -n "checkbox|\\[x\\]|Reply in either format|Any additional skills" skills/role-creator/SKILL.md
```

Expected:
- Missing one or more required interaction lines.

**Step 2: Write minimal doc changes**

Update `skills/role-creator/SKILL.md` in `Required Workflow` and examples:
- Declare checkbox-first, numeric fallback behavior.
- Add a reusable recommendation template:

```text
Recommended skills:
1. [ ] skill-a
2. [ ] skill-b

Reply in either format:
- Edit checkboxes to [x]/[ ]
- Or send numbers, e.g. 1,2
```

- Add mandatory follow-up prompt for manual additions:

```text
Any additional skills to add manually? (comma-separated, optional)
```

- Add precedence rule:
  - If both checkbox and numeric forms are provided, checkbox result wins.
- Add explicit note:
  - Empty final skills list is allowed.

**Step 3: Verify doc changes**

Run:

```bash
rg -n "checkbox|\\[x\\]|Reply in either format|Any additional skills|Empty final skills list is allowed" skills/role-creator/SKILL.md
```

Expected:
- All required lines present.

**Step 4: Commit**

```bash
git add skills/role-creator/SKILL.md
git commit -m "docs: define checkbox-first skill selection flow for role-creator"
```

### Task 3: Final verification and OpenSpec sync

**Files:**
- Modify: `openspec/changes/add-role-creator-skill-factory/tasks.md` (if task status needs update)

**Step 1: Run full verification**

```bash
python3 -m unittest skills/role-creator/tests/test_create_role_skill.py -v
openspec validate add-role-creator-skill-factory --strict
```

Expected:
- Unit tests pass.
- OpenSpec validation passes.

**Step 2: Update task checklist**

- Ensure all completed tasks are marked `- [x]` and reflect reality.

**Step 3: Commit verification artifacts (if changed)**

```bash
git add openspec/changes/add-role-creator-skill-factory/tasks.md
git commit -m "docs: update openspec task status for role-creator selection UX"
```
