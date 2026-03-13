# Workflow Skill Brainstorming

- Date: 2026-03-13
- Role used for brainstorming: `general strategist`
- Status: Approved design, no implementation started

## Problem Statement

The project needs a controller-facing workflow skill that can:

1. Create a workflow configuration file in the current project directory.
2. Execute that workflow as the main controller by orchestrating existing `agent-team` capabilities.

The workflow must support delivery flows such as:

- `controller -> cto breakdown -> controller dispatch -> development -> testing`
- `controller -> cto breakdown -> test-first authoring -> development -> testing verification`

The workflow skill must not require any modification to `agent-team` itself. It should operate entirely as a controller-side orchestration layer that composes existing `agent-team` commands.

## Goals

- Provide a dedicated `workflow` skill for the main controller.
- Support `create workflow` and `run workflow <file>` as the two core user-facing actions.
- Generate a reusable workflow template file in the current project.
- Execute approved workflows with limited branching and controller confirmation points.
- Reuse existing `agent-team` commands for worker lifecycle and communication.
- Persist runtime state separately from the workflow template so runs can pause and resume safely.

## Non-Goals

- Do not modify `agent-team` commands, protocol, config structure, or source code.
- Do not build a generic workflow engine with arbitrary scripting or unrestricted graph semantics.
- Do not generate implementation code as part of this brainstorming phase.
- Do not bypass brainstorming before creating workflow templates.

## Constraints And Assumptions

- The skill is primarily for the main controller, not for worker roles.
- Workflow configuration should describe role-level flow, while runtime resolves role-to-worker mappings.
- Some nodes must be able to pause for controller confirmation.
- Branching is limited and explicit, not an open-ended expression language.
- Existing `agent-team` primitives such as `worker create`, `worker open`, `worker assign`, `reply-main`, and `worker merge` remain the only execution backend.

## Candidate Approaches

### Approach A: Lightweight Documentation Workflow

The skill only generates `workflow.yaml`. The controller manually interprets and runs each step with `agent-team`.

Trade-offs:

- Pros: Lowest implementation cost and minimal coordination logic.
- Cons: Too much manual work, weak automation, poor fit for the stated goal.

### Approach B: Skill-Driven Controller Orchestration (Recommended)

The skill owns both workflow creation and workflow execution. It generates an approved template, then runs that template by orchestrating existing `agent-team` commands and persisting runtime state.

Trade-offs:

- Pros: Best match for the requested outcome, preserves `agent-team` boundaries, supports automation and resumability.
- Cons: Requires runtime state handling, node progression logic, and controller confirmation semantics.

### Approach C: Native `agent-team workflow` Commands

Workflow becomes a first-class CLI feature inside `agent-team`.

Trade-offs:

- Pros: Strong long-term cohesion and native UX.
- Cons: Highest implementation cost and explicitly conflicts with the requirement to avoid modifying `agent-team`.

## Recommended Design

Choose Approach B.

Build a controller-only `workflow` skill that sits above `agent-team`, using it as a stable execution backend. The skill is responsible for:

- Guided workflow design before file creation.
- Workflow template generation.
- Runtime orchestration.
- Run-state persistence and resumption.
- Limited branch selection and confirmation gating.

## Architecture

The design is split into two layers:

- `workflow.yaml`
  A reusable, approved workflow template that captures flow intent.
- `run-state.yaml`
  A per-execution state record that captures runtime progress and decisions.

The `workflow` skill exposes two primary actions:

- `create workflow`
- `run workflow <file>`

Execution responsibilities:

- The skill reads a workflow template.
- It resolves roles to workers at runtime.
- It decides whether to create or reuse workers.
- It dispatches tasks to workers through `agent-team`.
- It waits for worker feedback.
- It pauses when controller confirmation is required.
- It advances, branches, retries, or blocks based on the workflow rules.

## Components

### 1. Workflow Template

Stored in the current project directory and defines:

- workflow name
- participating roles
- nodes
- limited branches
- controller confirmation requirements
- default execution behavior

### 2. Controller-Only Workflow Skill

Used only by the main controller. It provides:

- `create workflow`
- `run workflow <file>`

It performs orchestration and command composition only. It does not require any change to `agent-team`.

### 3. Run State File

A runtime state artifact maintained by the workflow skill. It records:

- run id
- current node
- role-to-worker mapping
- node results
- blocking reason
- pending controller approvals

### 4. Agent-Team Command Bridge

This is a logical adapter layer inside the skill, not a modification to `agent-team`. It invokes existing commands such as:

- `agent-team worker create`
- `agent-team worker open`
- `agent-team worker assign`
- `agent-team worker status`
- `agent-team reply`
- `agent-team worker merge`

## Data Flow

### Create Workflow

Workflow creation must follow a brainstorming-first process similar to `skills/role-creator/SKILL.md`:

1. Start with structured brainstorming.
2. Clarify workflow purpose, participating roles, branching mode, approval points, and rollback rules.
3. Present candidate workflow options.
4. Present a concise approved design.
5. Only after approval, generate the workflow configuration file.

This means workflow file generation is gated behind explicit design approval.

### Run Workflow

1. The controller runs `run workflow <file>`.
2. The skill validates the workflow template.
3. The skill initializes a run-state file.
4. It resolves required roles and maps them to workers.
5. It starts from the entry node and advances node by node.
6. It pauses on nodes that require controller confirmation.
7. It records all runtime decisions and outcomes in run-state.

### Example Minimal Template Shape

```yaml
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

  - id: qa_write_tests
    type: assign_role_task
    actor: qa
    task: 先编写验收测试或测试用例
    next: dev_implement

  - id: dev_implement
    type: assign_role_task
    actor: dev
    task: 根据需求与测试实现功能
    next: qa_verify

  - id: qa_verify
    type: verify_or_test
    actor: qa
    task: 执行验证并反馈结果
    branches:
      - when: passed
        next: controller_finish
      - when: failed
        next: dev_fix

  - id: dev_fix
    type: assign_role_task
    actor: dev
    task: 根据测试反馈修复问题
    next: qa_verify

  - id: controller_finish
    type: controller_task
    requires_confirmation: true
    task: 决定是否合并与关闭流程
    end: true
```

## Node Types

The first version should support a small fixed set of node types:

- `controller_task`
- `assign_role_task`
- `wait_for_completion`
- `decision`
- `handoff`
- `verify_or_test`
- `merge`

This is enough to model both `dev-first` and `test-first` flows without building a generic execution engine.

## Error Handling

### Configuration Errors

Examples:

- missing required fields
- undefined roles
- missing node references
- invalid branch targets

Handling:

- fail fast before execution starts
- report exact field or node id causing the problem

### Role Or Worker Preparation Failures

Examples:

- role does not exist
- worker cannot be created
- worker session cannot be opened

Handling:

- mark run as `blocked`
- let controller choose retry, remap, or terminate

### Node Execution Failures

Two categories:

- `system failure`
  command failure, missing session, unreadable state
- `business failure`
  verification/test node reports failure

Handling:

- system failure pauses the run for controller action
- business failure follows the workflow branch, for example back to `dev_fix`

### Confirmation Pauses

When a node has `requires_confirmation: true`:

- persist run state as waiting for confirmation
- allow later resume without restarting the workflow

### Uninterpretable Worker Feedback

If worker output cannot be mapped to the expected result:

- do not auto-advance
- mark the node as needing manual interpretation
- let controller choose the next step

### Idempotency And Recovery

- completed nodes should not re-run unless the controller explicitly retries them
- state restoration must be safe across pauses and restarts
- duplicate assignment of the same node should be prevented

## Validation And Test Strategy

Validation should happen in four layers.

### 1. Template Validation

Check:

- required fields
- unique node ids
- valid `next` references
- valid `branches.next` references
- valid actor-to-role mappings

### 2. Pre-Run Validation

Check:

- referenced roles exist
- `agent-team` is available in the environment
- existing worker reuse or auto-create policy is satisfiable

### 3. Runtime Validation

Check after every node:

- state is written successfully
- node result is complete
- resume position is valid
- repeated dispatch risk is not introduced

### 4. Golden Workflow Samples

At minimum, validate two canonical flows:

- `dev-first`
  `cto breakdown -> dev implement -> qa verify -> controller finish`
- `test-first`
  `cto breakdown -> qa write tests -> dev implement -> qa verify -> controller finish`

## Risks And Mitigations

### Risk: Workflow DSL grows too complex

Mitigation:

- limit branching to explicit predefined cases
- avoid arbitrary expressions and open-ended transitions

### Risk: Worker feedback is inconsistent

Mitigation:

- define expected outcome mapping clearly
- pause for controller review on ambiguous responses

### Risk: Runtime state drifts from actual worker progress

Mitigation:

- persist state after every node
- require explicit result transitions
- make recovery and retry controller-visible

### Risk: Controller automation crosses `agent-team` boundaries

Mitigation:

- treat `agent-team` as a fixed backend
- only compose public commands
- do not depend on internal implementation details

## Open Questions

- Whether workflow templates should default to a specific directory such as `.agents/workflows/` or another project-local path.
- How run-state files should be named and organized for multiple concurrent workflow runs.
- Whether future versions should add optional parallel branches after the initial limited-branch release.

## Final Approved Direction

The approved direction is:

- Build a controller-only `workflow` skill.
- Require brainstorming and approval before generating workflow files.
- Generate a project-local workflow configuration file.
- Run workflows by orchestrating existing `agent-team` commands only.
- Keep template files and runtime state files separate.
- Support limited branching with controller confirmation points.
- Prioritize `dev-first` and `test-first` delivery workflows in the initial design.
