# Brainstorming Process

## Anti-Pattern: "This Is Too Simple To Need A Design"

Every assignment goes through this process. A single-file fix, a config change, a one-liner — all of them. "Simple" tasks are where unexamined assumptions cause the most wasted work. The design can be short (a few sentences for truly simple tasks), but you MUST present it and get approval.

## Checklist

When the user intends to assign new work to a worker, complete these steps in order:

1. **Explore project context** — check the role's `system.md`, existing `openspec/specs/`, project files, docs, and recent commits
2. **Ask clarifying questions** — one at a time, one question per message; prefer multiple choice when possible; focus on purpose, constraints, and success criteria
3. **Propose 2-3 approaches** — with trade-offs; lead with your recommendation and reasoning
4. **Present design** — scale each section to its complexity (a few sentences if straightforward, up to 200-300 words if nuanced); ask after each section whether it looks right; cover: architecture, components, data flow, error handling, testing as relevant
5. **User approves design** — get explicit approval; be ready to revise if needed
6. **Write design doc** — save to `docs/brainstorming/YYYY-MM-DD-<topic>.md` and commit
7. **Write proposal** — save the approved design to a temp file
8. **Execute assign** — run `agent-team worker assign <worker-id> "<desc>" --design docs/brainstorming/<file>.md --proposal <file>`

## Key Principles

- **One question at a time** — don't overwhelm with multiple questions in one message
- **Multiple choice preferred** — easier to answer than open-ended when possible
- **YAGNI ruthlessly** — remove unnecessary features from all designs
- **Explore alternatives** — always propose 2-3 approaches before settling
- **Incremental validation** — present design section by section, get approval before moving on
- **Be flexible** — go back and clarify when something doesn't make sense

## Skip Conditions

Brainstorming can be skipped ONLY when:

- User explicitly says "just assign" or "直接分配"
- User provides a complete, detailed design document
