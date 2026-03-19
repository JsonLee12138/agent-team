# Strategies

## Overview

Strategic compact supports three policy levels:

- `light`
- `standard`
- `deep`

In v1, only `light` and `standard` are expected execution paths. `deep` is a reserved escalation contract.

## `light`

Use for cheap, fast context protection.

### Best for

- phase transition with limited state change
- pre-large-read before diff/log/test output
- quick resume when the controller still knows the task well

### Collection budget

- current goal
- current phase
- done so far
- next step
- immediate blockers/risks

### Expected outcome

A short recovery anchor that lets main survive the next read without rebuilding history.

## `standard`

Use for normal controller maintenance when context is drifting.

### Best for

- after several workflow transitions
- after multiple worker replies or review loops
- before a new dispatch/review cycle
- when resumed work needs a stable controller summary

### Collection budget

Use the full minimum schema from [state-collector.md](state-collector.md).

### Expected outcome

A reliable controller-side recovery packet that can restore Goal / Phase / Constraints / Done / Next / Risks.

## `deep`

Reserved for escalation only.

### Upgrade conditions

Only consider `deep` when one or more of these are true:

- `standard` cannot form a reliable recovery contract
- the controller must bridge several unresolved branches or blocked states
- recent decisions are too fragmented to summarize from bounded state
- the run has accumulated enough ambiguity that a small recovery packet would be misleading

### v1 rule

Define the contract only. Do not make `deep` the default or routine path.

## Selection Heuristic

Choose the smallest mode that preserves recovery quality:

1. Start with `light` for pre-large-read or obvious phase edge checkpoints.
2. Escalate to `standard` when the controller needs full current state, not just a short anchor.
3. Escalate to `deep` only when `standard` cannot recover the run safely.
