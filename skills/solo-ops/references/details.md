# solo-ops Reference Details

## Role directory layout

```
.worktrees/<name>/
  CLAUDE.md                          ← auto-generated from prompt.md on open
  agents/teams/<name>/
    config.yaml                      ← name, default_provider, pane_id
    prompt.md                        ← role system prompt (edit manually)
    tasks/
      pending/<timestamp>-<slug>.md  ← active tasks
      done/<timestamp>-<slug>.md     ← completed/archived tasks
```

## Task file format

Tasks are Markdown files. When a role completes a task, move the file from `tasks/pending/` to `tasks/done/`.

## Bidirectional communication

Role asks a question:
```bash
ask claude "<rolename>: <question>"
```

Main controller replies:
```bash
solo-ops reply <rolename> "<answer>"
```

Reply appears in the role's WezTerm tab as `[Main Controller Reply]`. The role AI must NOT proceed on blocked tasks until it receives a reply.

The `prompt.md` template includes this communication protocol automatically.
