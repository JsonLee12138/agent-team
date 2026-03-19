# Context Management Rules

## Trigger

Apply this rule whenever context grows, the task changes phase, or a provider session is degrading.

## Compact Triggers

1. MUST compact before starting a new logical phase after finishing the current one.
2. MUST compact before reading or pasting large outputs, logs, or diffs that are not yet necessary.
3. MUST compact when the active thread can no longer hold the task goal, constraints, and next actions clearly.
4. MUST compact before handoff, provider switch, session restart, or resumed work after a long pause.

## Provider Handling

- Claude MUST use `/compact` when any trigger above fires.
- Codex and Gemini MUST create a manual summary of goal, constraints, changed files, verification state, and next step.
