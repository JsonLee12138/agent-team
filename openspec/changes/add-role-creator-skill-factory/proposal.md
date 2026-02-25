# Change: Add `role-creator` Skill Factory

## Why
This repository currently ships a single skill (`agent-team`) and has no standardized way to scaffold additional role-specific skills. Users need a deterministic factory flow to create role skills (for example frontend/backend/product roles) with clear role boundaries, reusable metadata, and a consistent file contract.

## What Changes
- Add a new `role-creator` skill under `skills/role-creator`.
- Define an interactive creation workflow that:
  - gathers role intent and system prompt inputs,
  - runs `find-skills` first for AI recommendations,
  - lets users keep/remove/add skills before generation.
- Define output contract for generated role skill packages under `skills/<role-name>/` with:
  - `SKILL.md`
  - `references/role.yaml`
  - `system.md`
- Enforce role directory naming as English kebab-case.
- Define existing-directory handling as confirm + backup (`skills/.backup/<role-name>-<timestamp>/`) before overwrite.

## Impact
- Affected specs:
  - `role-creator-workflow` (new capability)
  - `generated-role-skill-contract` (new capability)
- Affected code:
  - New skill package in `skills/role-creator/`
  - New Python script and templates for deterministic file generation
  - Optional tests for generation and validation behavior
