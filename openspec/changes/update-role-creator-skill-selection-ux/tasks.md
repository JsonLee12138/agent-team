## 1. Role-creator interaction contract
- [x] 1.1 Update `skills/role-creator/SKILL.md` to show recommended skills as numbered checkboxes with explicit reply formats.
- [x] 1.2 Document numeric fallback (`1,3,5`) and mixed-input precedence rule (checkbox result wins).
- [x] 1.3 Add mandatory manual-additions prompt and explicit acceptance of empty final skills.
- [x] 1.4 Add one minimal retry example for unparseable selection replies.

## 2. Generator and template behavior
- [x] 2.1 Update `resolve_final_skills(...)` in `skills/role-creator/scripts/create_role_skill.py` to allow empty final output.
- [x] 2.2 Keep add/remove/dedupe behavior deterministic for non-empty selections.
- [x] 2.3 Update `skills/role-creator/assets/templates/role.yaml.tmpl` rendering path to emit `skills: []` for empty skills.

## 3. Test coverage and verification
- [x] 3.1 Add tests for checkbox-mode parsing behavior.
- [x] 3.2 Add tests for numeric-mode parsing behavior and mixed-input precedence.
- [x] 3.3 Add tests proving empty final skills are accepted and serialized as `skills: []`.
- [x] 3.4 Run `python3 -m unittest skills/role-creator/tests/test_create_role_skill.py -v`.
- [x] 3.5 Run `openspec validate update-role-creator-skill-selection-ux --strict`.

## 4. Dependencies and parallelism
- [x] 4.1 Complete Section 1 before finalizing Section 3 assertions for UX text expectations.
- [x] 4.2 Section 2.1 and 2.3 should be completed before Section 3.3.
- [x] 4.3 Sections 3.1 and 3.2 can be implemented in parallel once selection parsing behavior is finalized.
