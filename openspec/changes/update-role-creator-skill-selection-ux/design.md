## Context
`role-creator` already supports recommendation-driven skill curation, but its current UX instructions are not explicit enough for deterministic multi-select parsing in chat. The current generator also enforces at least one final skill, which prevents valid role packages with intentionally empty skills.

This change updates the selection interaction contract and keeps generation deterministic while permitting `skills: []`.

## Goals / Non-Goals
- Goals:
  - Define checkbox-first recommendation selection with numeric fallback.
  - Define deterministic precedence when checkbox and numeric formats are both present.
  - Keep manual-skill additions in the same flow and de-duplicate merged skills.
  - Allow an empty final skills list and serialize it as `skills: []`.
- Non-Goals:
  - No redesign of role metadata collection (name/description/scope/system goal).
  - No changes to role naming validation or backup/overwrite behavior.
  - No changes to the three-file output contract (`SKILL.md`, `role.yaml`, `system.md`).

## Decisions
- Decision: Keep selection parsing rules in the role-creator workflow contract and testable script helpers.
  - Why: The user-facing flow is conversational, but deterministic parsing behavior still needs code-level test coverage.
- Decision: Checkbox format is the primary source of truth when both checkbox and numeric forms are provided.
  - Why: Checkbox replies encode explicit include/exclude state per option and are less ambiguous than index-only input.
- Decision: Preserve manual additions as a mandatory follow-up prompt after recommendation parsing.
  - Why: Recommendation quality may vary; users need a predictable place to add missing skills.
- Decision: Treat empty final skill list as valid output and serialize explicitly as `skills: []`.
  - Why: Empty role skills is a valid user intent and must not collapse to null/implicit YAML values.

## Alternatives Considered
- Numeric-only selection:
  - Rejected because it is less expressive for explicit unselection and higher risk for ambiguity.
- Blocking empty skills and forcing at least one manual skill:
  - Rejected because it conflicts with requested behavior and limits role definitions that intentionally avoid auto-mounted skills.

## Risks / Trade-offs
- Risk: More parsing paths increase input-handling edge cases.
  - Mitigation: Add targeted tests for checkbox mode, numeric mode, mixed mode precedence, and unparseable input retries.
- Risk: Empty list serialization may regress to `skills:` (implicit null) in template changes.
  - Mitigation: Use an explicit rendered field and enforce `skills: []` via unit test.

## Migration Plan
1. Update OpenSpec deltas for workflow and contract capabilities.
2. Implement role-creator UX copy updates in `SKILL.md`.
3. Implement generator/template updates for empty skills support.
4. Add/adjust tests and run strict OpenSpec + unit validation.

## Open Questions
- None blocking. The design input defines selection precedence, fallback behavior, and empty-skill contract explicitly.
