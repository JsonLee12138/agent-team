# Merge Workflow Rules

## Trigger

Apply this rule when preparing a worker branch for assignment, merge, or controller-side synchronization.

## Controller-Side Synchronization

- MUST keep worker-side sessions free of `git rebase` and `git merge` inside the worker worktree.
- MUST perform any required rebase from the controller side before assignment when the worker is idle.
- MUST stop and surface conflicts immediately if controller-side rebase fails.

## Merge Safety

- MUST merge worker branches back through the controller workflow.
- MUST review generated files and ignore-only artifacts before merge so worker-local prompts or metadata do not enter commits.
- MUST preserve the repository's tracked deliverables while excluding controller-managed bootstrap files.
