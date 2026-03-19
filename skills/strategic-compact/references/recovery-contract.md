# Recovery Contract

## Purpose

After compact finishes, the controller should be able to resume work from a minimal recovery packet instead of re-reading the whole transcript.

## Required Anchors

Every successful recovery contract should preserve these anchors:

- `Goal`
- `Phase`
- `Constraints`
- `Done`
- `Next`
- `Risks`

## Recommended Shape

Use concise controller-facing bullets.

```text
Goal: <current objective>
Phase: <current phase or workflow node>
Constraints:
- <constraint>
Done:
- <completed item>
Next:
- <next action>
Risks:
- <risk or blocker>
```

## Field Guidance

### Goal

The immediate controller objective, not the whole project mission.

### Phase

Prefer an explicit workflow or task phase label.
Examples:
- `workflow: qa_verify`
- `reviewing worker completion`
- `pre-large-read before test output`

### Constraints

Include only constraints that can change the next decision.
Examples:
- main-first compact policy
- worker completion chain remains verify -> archive -> reply-main
- no automatic merge/delete

### Done

List only already-resolved items that matter for the next step.

### Next

State the next controller action as a concrete step.
Examples:
- `read worker diff, then decide review vs merge`
- `run workflow state wait and pause for worker reply`

### Risks

List blockers, uncertainty, or missing evidence.
If there are none, say so briefly.

## Fallback Rule

Do not generate a generic handoff summary by default.
Use fallback wording only when the normal recovery contract cannot be formed from bounded state.

Example fallback:

```text
Risks:
- bounded state is insufficient; manual reread of the latest worker reply is required
```
