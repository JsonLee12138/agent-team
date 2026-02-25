## ADDED Requirements

### Requirement: Generator MUST Accept Empty Confirmed Skills List
The role generator SHALL treat an empty confirmed skills list as valid input and continue file generation.

#### Scenario: Final skills list is empty
- **WHEN** the workflow provides an empty skills list to the generator
- **THEN** generation succeeds without skill-related validation errors
- **AND** managed files are still produced under `skills/<role-name>/`

### Requirement: `role.yaml` MUST Serialize Empty Skills as `skills: []`
Generated `role.yaml` SHALL encode empty skills as an explicit YAML empty array (`skills: []`) rather than implicit null or omitted field output.

#### Scenario: Empty skills are written to metadata
- **WHEN** generation runs with no skills
- **THEN** `role.yaml` contains the exact field form `skills: []`
- **AND** downstream consumers can parse skills as an empty list deterministically

### Requirement: Non-Empty Skills Serialization MUST Remain List-Shaped
For non-empty selections, generated `role.yaml` SHALL continue serializing skills as a YAML list of string entries.

#### Scenario: One or more skills are provided
- **WHEN** generation runs with selected skills
- **THEN** `role.yaml` contains a `skills:` field with list items
- **AND** each entry is serialized as a string value in stable order
