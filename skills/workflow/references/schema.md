# Workflow Schema

## Workflow Template

Use a project-local YAML file with this shape:

```yaml
version: 1
name: feature-delivery
roles:
  cto:
    role: cto
  dev:
    role: vite-react-dev
  qa:
    role: qa
defaults:
  execution_mode: semi_auto
  create_worker_if_missing: true
entry: cto_breakdown
nodes:
  - id: cto_breakdown
    type: assign_role_task
    actor: cto
    task: 拆分需求并给出实施子任务
    next: controller_dispatch

  - id: controller_dispatch
    type: controller_task
    requires_confirmation: true
    task: 审阅拆分结果并决定分发路径
    branches:
      - when: test_first
        next: qa_write_tests
      - when: dev_first
        next: dev_implement
```

Required top-level fields:

- `version`: integer, start with `1`
- `name`: workflow name
- `roles`: map of workflow aliases to role definitions
- `entry`: first node id
- `nodes`: ordered list of nodes

Recommended top-level fields:

- `defaults.execution_mode`: `semi_auto` or `manual`
- `defaults.create_worker_if_missing`: boolean
- `defaults.reuse_worker`: boolean

Role shape:

```yaml
roles:
  dev:
    role: vite-react-dev
    worker: vite-react-dev-001
```

- `role`: required role name
- `worker`: optional preferred worker id for reuse

## Node Types

Use a small fixed node set:

- `controller_task`
- `assign_role_task`
- `wait_for_completion`
- `decision`
- `handoff`
- `verify_or_test`
- `merge`

Common node fields:

- `id`: required unique node id
- `type`: required supported node type
- `task`: recommended human-readable instruction
- `actor`: required for role-backed nodes
- `next`: optional next node id for linear flow
- `branches`: optional list of `{ when, next }`
- `requires_confirmation`: optional boolean
- `end`: optional boolean terminal marker

Branch rules:

- Use explicit branch labels only.
- Every `branches[].next` target must exist.
- Use either `next`, `branches`, or `end: true`.
- Keep branching shallow and controller-visible.

## Run-State File

Store one run-state file per execution:

```yaml
version: 1
run_id: 20260313-153000-feature-delivery
workflow_file: .agents/workflow/feature-delivery.yaml
workflow_name: feature-delivery
status: ready
current_node: cto_breakdown
pending_confirmation: null
blocking_reason: null
role_worker_map: {}
node_states:
  cto_breakdown:
    status: pending
decision_log: []
created_at: "2026-03-13T15:30:00Z"
updated_at: "2026-03-13T15:30:00Z"
```

Run-state status values:

- `ready`
- `running`
- `waiting`
- `waiting_confirmation`
- `blocked`
- `completed`

Per-node status values:

- `pending`
- `running`
- `waiting`
- `completed`
- `blocked`

## Outcome Mapping

Use workflow-native outcomes to move through branches:

- controller dispatch: `dev_first`, `test_first`
- verification: `passed`, `failed`
- confirmation gate: any branch label defined on the node

If worker feedback cannot be mapped to one of the allowed outcomes:

- do not auto-advance
- mark the run as `blocked`
- record the ambiguous summary in `decision_log`
