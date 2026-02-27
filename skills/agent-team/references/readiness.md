# Assign Readiness Protocol

## Scope

This protocol is mandatory before every `agent-team worker assign` dispatch.

- Applies to all workers and all providers
- Prompt-level guardrail only (no CLI enforcement in this phase)
- Timing constants: first wait `5s`, retry interval `5s`, max attempts `3`

## Required Handshake Tokens

- Ping: `AGENT_TEAM_PING <worker-id> <attempt>`
- Pong: `AGENT_TEAM_PONG <worker-id> <attempt>`

The controller MUST validate that `<worker-id>` and `<attempt>` match exactly.

## State Flow

1. Send ping attempt `1`
2. Wait `5s` for matching pong
3. If pong arrives, continue to assign
4. If timeout, inspect worker window and classify state
5. If provider workspace is not open (`codex`, `claude`, or `opencode`), send provider open command
6. Wait `5s`, then send next ping attempt
7. Retry until attempt `3`
8. If no valid pong by attempt `3`, throw structured error and stop assign

## Window Diagnosis Checklist (On Timeout)

On every timeout, the controller MUST read the worker window and capture observations.

- Is the worker session alive?
- Is the target provider workspace open (`codex` / `claude` / `opencode`)?
- Is the worker blocked on auth, permission, login, or command failure?
- Is there any obvious crash, hang, or unresponsive state?

If workspace is not open, send the provider open command first, then continue retry flow.

## Retry Policy

- Attempt budget: `1..3`
- Timeout per attempt: `5s`
- Delay before next attempt: `5s` after diagnosis/open action
- No silent retries: each retry must include diagnosis result

## Error Categories

Use these canonical error types:

- `no_pong_timeout`
- `workspace_not_open`
- `provider_launch_failed`
- `unknown_worker_state`

## Structured Failure Output

When readiness fails at attempt 3, output a structured failure block and stop assign.

Required fields:

- `worker-id`
- `attempt` (`1|2|3`)
- `error_type`
- `window_inspected` (`true|false`)
- `open_command_sent` (`true|false`)
- `observation` (short window diagnosis)
- `action` (`assign_stopped`)

Example:

```text
[Assign Readiness Error]
worker-id: frontend-dev-001
attempt: 3
error_type: no_pong_timeout
window_inspected: true
open_command_sent: true
observation: provider pane opened but no AGENT_TEAM_PONG received in all retries
action: assign_stopped
```

## Controller Rules (MUST)

- MUST run this protocol before every assign dispatch
- MUST stop assign after third failed attempt
- MUST provide structured failure output
- MUST NOT skip window diagnosis on timeout
- MUST NOT dispatch assign without successful pong
