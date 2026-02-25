## ADDED Requirements

### Requirement: AI Recommendation MUST Precede Skill Selection
The role creation workflow SHALL execute a `find-skills` recommendation step before finalizing the role skill list for generation.

#### Scenario: Recommendation step runs before confirmation
- **WHEN** a user starts creating a new role skill
- **THEN** the workflow runs `find-skills` to produce candidate skills before file generation

#### Scenario: Recommendation returns no usable result
- **WHEN** `find-skills` is unavailable or returns an empty set
- **THEN** the workflow asks the user to provide skills manually and continues

### Requirement: User MUST Confirm Final Skills List
The workflow SHALL present candidate skills and allow the user to keep, remove, or add entries before generation.

#### Scenario: User edits recommendations
- **WHEN** recommended skills are displayed
- **THEN** the user can remove irrelevant items and add missing items
- **AND** generation uses only the final user-confirmed list

### Requirement: Role Name MUST Be Kebab-Case
The generator SHALL enforce English kebab-case naming for `<role-name>` target directories under `skills/`.

#### Scenario: Invalid role name is provided
- **WHEN** a user inputs a role name that is not kebab-case
- **THEN** generation is blocked
- **AND** the tool provides a suggested normalized kebab-case name

### Requirement: Existing Role Directory MUST Be Backed Up Before Overwrite
If `skills/<role-name>/` already exists, the workflow SHALL require explicit user confirmation and create a backup before overwriting generated files.

#### Scenario: User confirms overwrite
- **WHEN** target role directory already exists and user confirms overwrite
- **THEN** the existing directory is backed up to `skills/.backup/<role-name>-<timestamp>/`
- **AND** generation proceeds to overwrite managed output files

#### Scenario: User declines overwrite
- **WHEN** target role directory already exists and user declines overwrite
- **THEN** generation stops without modifying existing files
