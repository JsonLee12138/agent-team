# Role Creator Skill Selection UX Design

## Context
`skills/role-creator/SKILL.md` currently describes keep/remove/add selection, but the interaction copy is not explicit enough for a reliable multi-select flow in chat. The requested behavior is:
- show recommended skills as a selectable list,
- support multi-select from users,
- allow manual skill input in the same flow,
- allow final skills to be empty, persisted as `skills: []`.

## Goals
- Make the selection UX explicit and deterministic in `SKILL.md`.
- Use checkbox-first interaction, with index-based selection as fallback.
- Keep compatibility with existing generator script invocation flow.
- Allow empty final skills list without blocking generation.

## Non-Goals
- No role creation flow redesign outside skill selection UX.
- No changes to role naming rules, backup/overwrite policy, or file contract.
- No changes to OpenSpec requirements in this design step.

## Decision Summary
1. Selection mode:
   - Primary: checkbox editing (`[x]` / `[ ]`).
   - Secondary fallback: numeric selection (`1,3,5`).
2. Parse precedence:
   - If both checkbox and numeric forms are present, checkbox result wins.
3. Manual additions:
   - Always ask an extra question for user-entered skills (comma-separated).
   - Merge with selected list and de-duplicate.
4. Empty list behavior:
   - Final skills may be empty.
   - Persist as `skills: []` in generated `role.yaml`.

## Interaction Design

### Step A: Show recommendation options
Display `find-skills` recommendations as numbered checkboxes:

```text
Recommended skills:
1. [ ] ui-ux-pro-max
2. [ ] better-icons
3. [ ] vitest
4. [ ] vite

Reply in either format:
- Edit checkboxes to [x]/[ ]
- Or send numbers, e.g. 1,3,4
```

### Step B: Collect manual additions
After selection parsing:

```text
Any additional skills to add manually? (comma-separated, optional)
```

### Step C: Confirm final list
Echo the merged result and proceed. If empty:
- explicitly confirm empty list is accepted,
- continue generation.

## Error Handling
- `find-skills` returns empty/unavailable:
  - skip recommendation list,
  - ask only manual additions.
- Unparseable user reply:
  - show one minimal valid example and ask once again.
- Both checkbox and numeric formats present:
  - parse both, apply checkbox result, and inform user briefly.

## Testing Considerations
- Selection parsing tests for checkbox mode.
- Selection parsing tests for numeric mode.
- Conflict test (checkbox + numeric) ensures checkbox precedence.
- Empty final list test verifies role metadata uses `skills: []`.

## Impacted Files (Implementation Phase)
- `skills/role-creator/SKILL.md`
- `skills/role-creator/scripts/create_role_skill.py`
- `skills/role-creator/tests/test_create_role_skill.py`
- `skills/role-creator/assets/templates/role.yaml.tmpl`

## Acceptance Criteria
- `SKILL.md` contains explicit checkbox-first, numeric-fallback interaction instructions.
- `SKILL.md` includes mandatory manual-additions step.
- Generation flow allows empty final skills without error.
- Generated `role.yaml` serializes empty list as `skills: []`.
