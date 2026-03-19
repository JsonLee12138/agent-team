# Context Management Rules

## Trigger

Apply this rule whenever context grows, the task changes phase, or resumed work needs a recovery anchor.

## Mandatory Strategic-Compact Triggers

1. MUST enter strategic context management before starting a new logical phase after finishing the current one.
2. MUST enter strategic context management before reading or pasting large outputs, logs, replies, or diffs that may displace the current working context.
3. MUST enter strategic context management when the active thread can no longer hold the task goal, constraints, and next actions clearly.
4. MUST enter strategic context management before resumed work after a long pause, restart, or handoff.

## Required Entry

- When any trigger above fires, MUST use the `strategic-compact` skill.
- The skill decides the trigger type, recovery payload, and whether `light` or `standard` strategy is enough.
- Rules do not define provider-specific compact strategy branches.

## Policy Summary

- Main/controller is the default compact target.
- Worker compact is an exception path only for long-running, blocked, or multi-round worker tasks.
- Worker standard completion remains verify -> archive -> reply-main.
- Reuse existing `agent-team compact` transport instead of inventing a new compact path in rules.
