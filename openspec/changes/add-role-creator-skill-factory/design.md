## Context
The repository currently includes `agent-team` skill assets but lacks a reusable factory workflow to create new role-specific skills. The requested behavior requires both conversation-level guidance (AI recommendation + user confirmation) and deterministic file generation (stable output structure and constraints).

## Goals / Non-Goals
- Goals:
  - Add a reusable `role-creator` skill package.
  - Produce deterministic role skill outputs in `skills/<role-name>/`.
  - Enforce explicit role boundaries and single-role focus in generated artifacts.
  - Handle overwrite safely via backup before replacement.
- Non-Goals:
  - No changes to `agent-team` CLI commands.
  - No automatic worktree/session creation.
  - No provider-specific metadata generation (`agents/openai.yaml` is out of scope).

## Decisions
- Decision: Implement `role-creator` as a self-contained skill using Python.
  - Why: Aligns with skill-creator conventions for script-backed deterministic behavior without requiring Go CLI changes.
- Decision: Separate recommendation from generation.
  - Why: `find-skills` recommendation remains conversational and user-driven, while script remains deterministic and side-effect-scoped.
- Decision: Output contract is fixed to three files (`SKILL.md`, `references/role.yaml`, `system.md`).
  - Why: Minimal, explicit, and portable skill package that captures behavior + metadata + system prompt.
- Decision: Existing target directory requires explicit confirmation and backup-first overwrite.
  - Why: Prevents accidental loss and keeps previous versions recoverable.

## Alternatives Considered
- Integrate into `agent-team` Go commands:
  - Rejected because user requested self-contained skill path and no direct CLI integration requirement.
- Pure prompt-based generation (no script):
  - Rejected because deterministic output and repeatability are required.

## Risks / Trade-offs
- Risk: `find-skills` may be unavailable or return empty recommendations.
  - Mitigation: Fall back to manual user-provided skills list before generation.
- Risk: Template/schema drift across generated role packages.
  - Mitigation: Keep templates in one place (`skills/role-creator/assets/templates/`) and validate generated artifacts in tests.

## Migration Plan
1. Add OpenSpec change artifacts (this change).
2. Implement `skills/role-creator` package and script.
3. Add generation validation tests.
4. Verify output examples and OpenSpec strict validation.

## Open Questions
- None blocking for this proposal. Naming, overwrite strategy, and file contract were confirmed during design discussion.
