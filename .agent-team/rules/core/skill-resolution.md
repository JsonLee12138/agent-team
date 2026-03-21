# Skill Resolution Rules

## Trigger

Apply this rule whenever a role needs a skill that is missing locally at runtime.

## Required Flow

1. MUST run `find-skills` first to search for a matching skill.
2. MUST allow only project-level installation for runtime skill resolution.
3. MUST NOT trigger any global skill installation path from a worker/runtime flow.
4. If project-level installation fails, MUST print a warning with the failure reason and a suggested next step, then continue the current task.
5. After the task, MAY suggest regenerating the role with `agent-team role create` so the skill is declared in `references/role.yaml`.
