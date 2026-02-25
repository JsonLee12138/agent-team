## ADDED Requirements

### Requirement: Generated Role Skill Package MUST Use Fixed Three-File Contract
The role generator SHALL create role skill packages at `skills/<role-name>/` with exactly these managed files: `SKILL.md`, `references/role.yaml`, and `system.md`.

#### Scenario: New role skill generation succeeds
- **WHEN** generation completes for a valid role input
- **THEN** `skills/<role-name>/SKILL.md` exists
- **AND** `skills/<role-name>/references/role.yaml` exists
- **AND** `skills/<role-name>/system.md` exists

### Requirement: `references/role.yaml` MUST Contain Core Metadata and Prompt Link
Generated `references/role.yaml` SHALL include role identity metadata, role boundary fields, selected skills, and `system_prompt_file: system.md`.

#### Scenario: Metadata is written from confirmed inputs
- **WHEN** the user confirms role description, boundaries, and skills
- **THEN** `references/role.yaml` records those values
- **AND** `references/role.yaml` contains `system_prompt_file: system.md`

### Requirement: Generated `SKILL.md` MUST Enforce Single-Role Focus
Generated `SKILL.md` SHALL define the role's in-scope responsibilities, out-of-scope boundaries, and handoff behavior for out-of-scope tasks.

#### Scenario: Role receives out-of-scope request
- **WHEN** a request falls outside the generated role boundary
- **THEN** `SKILL.md` directs the agent to avoid implementation beyond scope
- **AND** `SKILL.md` includes a handoff/escalation instruction

### Requirement: `system.md` MUST Be the System Prompt Source of Truth
Generated `system.md` SHALL contain the role-specific system prompt content referenced by `references/role.yaml`.

#### Scenario: System prompt is loaded via metadata
- **WHEN** a consumer reads `references/role.yaml`
- **THEN** it can resolve `system_prompt_file: system.md`
- **AND** `system.md` contains the role prompt text to inject
