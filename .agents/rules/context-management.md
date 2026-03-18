# Context Management Rules

## Trigger

Apply this rule whenever context grows, the task changes phase, or a provider session is degrading.

## Compact Triggers

1. MUST compact before starting a new logical phase after finishing the current one.
2. MUST compact before reading or pasting large outputs, logs, or diffs that are not yet necessary.
3. MUST compact when the active thread can no longer hold the task goal, constraints, and next actions clearly.
4. MUST compact before handoff, provider switch, session restart, or resumed work after a long pause.
5. MUST compact after resolving a blocker if the recovery path would otherwise require re-reading large prior context.

## Provider Handling

- Claude MUST use `/compact` when any trigger above fires, then continue from the compacted summary.
- Codex MUST create a manual summary of goal, constraints, changed files, verification state, and next step, then continue in a fresh turn if needed.
- Gemini MUST create the same manual summary because Gemini CLI has no native `/compact`; start a fresh prompt with that summary when context pressure remains high.

## Manual Summary Contents

- MUST include the task goal, current status, pending decision or next action, and any required commands.
- MUST include changed files, verification already run, and verification still required.
- MUST exclude obsolete exploration, repeated logs, and low-signal command output.
