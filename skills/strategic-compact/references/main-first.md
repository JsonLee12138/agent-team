# Main-First Policy

## Core Policy

The main/controller session is the default object of context management.

Why:
- it holds orchestration state across workflow, review, dispatch, and decisions
- it is the session most likely to accumulate long-lived token cost
- worker sessions should usually end after task completion and handoff back to main

## Default Lifecycle

Prefer this path:

1. worker implements and verifies
2. worker archives and sends `reply-main`
3. main reviews, decides merge/follow-up, and performs any needed compact
4. worker is kept or deleted based on explicit controller/user choice

This keeps compact on main instead of making worker persistence the default memory mechanism.

## Worker Default Rule

Do not recommend worker compact on the normal completion path.

Normal worker path remains:
- verify
- archive
- reply-main

After that, main decides whether compact is needed before review, re-dispatch, or merge discussion.

## Worker Exception Contract

A worker may use strategic compact only in these exception cases:

1. **Long task**
   - the same worker must continue for a substantial multi-step effort
2. **Blocked task**
   - the worker must preserve the current state while waiting for a decision or dependency
3. **Multi-round modification loop**
   - the same worker must continue through several edit/verify/fix rounds without handing back yet

## Worker Exception Guardrails

Even in exception cases:

- prefer the smallest possible recovery payload
- compact only the worker that truly must persist
- return to main-first behavior as soon as the exception ends
- do not reinterpret worker compact as approval to auto-merge or auto-delete

## Anti-Pattern

Do not normalize this behavior:

- keep workers open by default
- compact workers after every routine completion
- treat worker compact as a replacement for `reply-main`
- move controller recovery responsibility into workers
