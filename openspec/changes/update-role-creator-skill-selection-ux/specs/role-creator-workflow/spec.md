## ADDED Requirements

### Requirement: Recommendation Selection MUST Be Presented as Numbered Checkboxes
The role creation workflow SHALL present `find-skills` recommendations as a numbered checkbox list (`[ ]` / `[x]`) before skill confirmation.

#### Scenario: Recommended skills are shown in checkbox format
- **WHEN** `find-skills` returns at least one recommended skill
- **THEN** the workflow displays each recommendation as a numbered checkbox option
- **AND** the workflow includes instructions for replying in checkbox or numeric format

### Requirement: Numeric Multi-Select MUST Be Supported as Fallback
The workflow SHALL accept numeric index replies (for example `1,3,5`) as a fallback selection format for recommended skills.

#### Scenario: User replies with numeric indices
- **WHEN** recommended skills are displayed and the user replies with comma-separated indices
- **THEN** the workflow resolves the selected recommendations by index
- **AND** uses the resolved set as the current selected skills

### Requirement: Checkbox Parsing MUST Take Precedence Over Numeric Parsing
If a user reply includes both checkbox edits and numeric indices, the workflow SHALL prioritize checkbox parsing results.

#### Scenario: Mixed checkbox and numeric reply is provided
- **WHEN** the same user reply contains checkbox state edits and numeric indices
- **THEN** the workflow parses both forms
- **AND** applies checkbox-derived selection as the final selection for that step
- **AND** informs the user briefly that checkbox precedence was applied

### Requirement: Manual Skill Additions MUST Always Be Collected
After recommendation selection is parsed, the workflow SHALL always ask for optional manual skill additions and merge them with selected skills using de-duplication.

#### Scenario: User adds manual skills
- **WHEN** selection parsing finishes
- **THEN** the workflow asks for optional manual skills in comma-separated format
- **AND** merges the provided skills with selected recommendations without duplicates

#### Scenario: User provides no manual additions
- **WHEN** selection parsing finishes and the user submits no manual skills
- **THEN** the workflow continues using only the selected recommendation set

### Requirement: Workflow MUST Allow Empty Final Skill Selection
The workflow SHALL allow the final merged skill list to be empty and continue generation using the contract in `generated-role-skill-contract`.

#### Scenario: No skills are selected or added
- **WHEN** the user deselects all recommendations and provides no manual additions
- **THEN** generation continues without blocking on missing skills
- **AND** the generated metadata records an explicit empty skills list

### Requirement: Unparseable Selection Input MUST Trigger a Single Retry Prompt
If a recommendation selection reply cannot be parsed as checkbox edits or numeric indices, the workflow SHALL return one minimal valid example and request one retry.

#### Scenario: Selection reply cannot be parsed
- **WHEN** a user selection reply does not match supported checkbox or numeric formats
- **THEN** the workflow responds with one concise valid example
- **AND** requests the user to retry selection once
