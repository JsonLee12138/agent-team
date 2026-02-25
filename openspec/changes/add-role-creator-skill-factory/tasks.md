## 1. Specification and scaffolding
- [x] 1.1 Create `skills/role-creator/` with `SKILL.md`, `scripts/`, and `assets/templates/`.
- [x] 1.2 Define workflow instructions in `skills/role-creator/SKILL.md` including:
  - run `find-skills` recommendations first,
  - user keep/remove/add confirmation,
  - deterministic generation invocation.

## 2. Deterministic generator
- [x] 2.1 Implement `skills/role-creator/scripts/create_role_skill.py` to generate:
  - `skills/<role-name>/SKILL.md`
  - `skills/<role-name>/references/role.yaml`
  - `skills/<role-name>/system.md`
- [x] 2.2 Enforce role-name kebab-case validation and provide normalization hint on invalid input.
- [x] 2.3 Implement existing-directory flow: confirm overwrite, backup to `skills/.backup/<role-name>-<timestamp>/`, then overwrite managed files.

## 3. Template and schema contract
- [x] 3.1 Add templates for generated files with placeholders for role name, description, boundaries, and selected skills.
- [x] 3.2 Ensure generated `SKILL.md` contains single-role-focus and out-of-scope handoff behavior.
- [x] 3.3 Ensure generated `references/role.yaml` includes required metadata and `system_prompt_file: system.md`.

## 4. Validation and tests
- [x] 4.1 Add tests for name validation and deterministic rendering.
- [x] 4.2 Add tests for overwrite + backup behavior and manual-skills fallback path.
- [x] 4.3 Run validation checks for generated outputs and document verification command(s).
