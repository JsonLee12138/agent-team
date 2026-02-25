# Change: Update Role Creator Skill Selection UX

## Why
`skills/role-creator/SKILL.md` currently describes keep/remove/add behavior, but it does not define an explicit multi-select interaction format for chat replies. The current generator logic also rejects empty final skill lists, which conflicts with the new design requirement to allow `skills: []`.

## What Changes
- Refine role-creator interaction contract to a deterministic selection flow:
  - show recommendation options as numbered checkboxes,
  - support numeric selection as fallback,
  - define checkbox parsing precedence when both formats appear in the same reply,
  - always ask for optional manual skill additions before final confirmation.
- Allow empty final skill selection and continue generation without blocking.
- Require empty skills to be serialized in `role.yaml` as explicit `skills: []`.
- Add validation coverage for checkbox/numeric selection behavior, precedence, and empty-skill serialization.

## Impact
- Affected specs:
  - `role-creator-workflow` (selection interaction behavior)
  - `generated-role-skill-contract` (empty skills serialization contract)
- Affected code:
  - `skills/role-creator/SKILL.md`
  - `skills/role-creator/scripts/create_role_skill.py`
  - `skills/role-creator/tests/test_create_role_skill.py`
  - `skills/role-creator/assets/templates/role.yaml.tmpl`
- Current gap noted during proposal prep:
  - `openspec/specs/` is currently empty, so this change remains as a proposal delta layered on top of existing completed change artifacts.
