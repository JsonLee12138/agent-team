---
name: brainstorming
description: "Use when users explicitly ask to brainstorm, shape requirements, compare approaches, or produce a planning/design document before implementation. Turn rough ideas into validated brainstorming/design docs through focused dialogue, role-based analysis, and explicit user approval. Do not trigger this skill for straightforward implementation requests that do not need dedicated design exploration."
---

# Brainstorming Ideas Into Designs

## Overview

Turn ideas into clear, approved brainstorming documents through collaborative dialogue.

Always support role-based brainstorming:
- If the user specifies a role, follow that role's perspective (for example: PM, frontend architect, backend architect, UX designer).
- If no role is specified, use the `general strategist` role.

<HARD-GATE>
Do NOT invoke any implementation skill, write any code, scaffold any project, or take any implementation action until you have presented the design and the user has approved it.
</HARD-GATE>

## Scope Rule

Use this process when the user wants dedicated design exploration, option comparison, or requirement shaping before implementation.

For straightforward implementation requests, small fixes, or clearly specified changes, do not force this full brainstorming flow.
When lightweight design help is enough, keep the output short and ask only the minimum questions needed to remove ambiguity.

## Checklist

Choose the lightest checklist that fits the request.

### Standard flow

1. Explore project context: inspect only the files, docs, and recent changes needed for the topic.
2. Ask clarifying questions: prefer one focused question at a time when ambiguity remains.
3. Propose 2-3 approaches when trade-offs matter; include a recommendation.
4. Present the design in concise sections covering only the dimensions that matter for the request.
5. When the topic is planning-oriented, explicitly ask which target object the brainstorming is for: `roadmap`, `milestone`, `phase`, `task`, or `generic topic`.
6. When relevant, ask whether the brainstorming should build on an existing object or reference set.
7. If the user wants a saved artifact, ask for save location choice before writing.
8. Stop after delivering the brainstorming output unless the user explicitly asks to continue.

### Lightweight flow

Use this lighter path for narrow requirement shaping:
1. Confirm the goal and key constraint.
2. Present the recommended approach first.
3. Mention alternatives only if they materially change scope, risk, or cost.
4. Deliver the result inline unless the user asks to save it.

## Process Flow

```dot
digraph brainstorming {
    "Explore project context" [shape=box];
    "Ask clarifying questions" [shape=box];
    "Propose 2-3 approaches" [shape=box];
    "Present design sections" [shape=box];
    "User approves design?" [shape=diamond];
    "Choose save location" [shape=box];
    "Write brainstorming doc" [shape=box];
    "Deliver doc and stop" [shape=doublecircle];

    "Explore project context" -> "Ask clarifying questions";
    "Ask clarifying questions" -> "Propose 2-3 approaches";
    "Propose 2-3 approaches" -> "Present design sections";
    "Present design sections" -> "User approves design?";
    "User approves design?" -> "Present design sections" [label="no, revise"];
    "User approves design?" -> "Choose save location" [label="yes"];
    "Choose save location" -> "Write brainstorming doc" [label="default / custom"];
    "Choose save location" -> "Deliver doc and stop" [label="skip saving"];
    "Write brainstorming doc" -> "Deliver doc and stop";
}
```

Terminal state is `Deliver doc and stop`.
Do not continue into implementation planning or implementation unless the user explicitly asks for the next step.

## Save Location Rule (Required)

Before any file write, always ask the user to choose where the brainstorming doc should live.

### Standard choices
1. Default directory: `docs/brainstorming/`
2. Custom directory: user-provided path
3. Skip saving: do not write any file

### Extra choice for planning-layer targets
If the brainstorming target is `roadmap`, `milestone`, or `phase`, also offer:
4. Target object directory: save the brainstorming doc next to the target object's primary file when the user explicitly wants the doc colocated with that object.

### Decision rule
- If the user has not asked for colocated storage, prefer `docs/brainstorming/` as the default.
- Only use the target object directory when the user explicitly wants the brainstorming doc attached to that roadmap/milestone/phase.
- Only use a custom directory when the user explicitly provides one.
- If the user chooses skip saving, do not write any file.

### Filename rule
If the user chooses default, custom, or target object directory, use:
- `<topic>-YYYY-MM-DD-brainstorming.md`

For the default directory, the full path becomes:
- `docs/brainstorming/<topic>-YYYY-MM-DD-brainstorming.md`

For a target object directory, keep the same filename pattern and place it in that object's directory.

Place the topic first so related brainstorming docs group together more naturally during directory browsing, with the date still preserved for chronology.

For a custom directory, write to the exact directory the user provides with the same filename pattern.

If the user chooses skip saving, do not write any file — output the doc content inline in the chat only.

## Questioning Rules

- Ask only one question per message.
- Prefer multiple-choice questions where possible.
- Focus on purpose, constraints, and success criteria.
- Go back and re-clarify when responses are ambiguous.

## Design Presentation Rules

- Present 2-3 approaches before finalizing only when meaningful trade-offs exist.
- Lead with the recommended option and explain why.
- Keep each section concise for simple tasks and detailed for complex tasks.
- For lightweight brainstorming, ask for a single overall approval instead of section-by-section approval.
- Use section-by-section approval only when the design is large, risky, or materially ambiguous.

## Doc Content Guidelines

Include:
- Role used for brainstorming (explicit role or `general strategist`)
- Target object (`roadmap`, `milestone`, `phase`, `task`, or `generic topic`) when relevant
- Problem statement and goals
- Constraints and assumptions
- Candidate approaches with trade-offs
- Recommended design
- For planning-layer topics, explicitly restate that `task` remains the only execution unit and that planning artifacts are planning/display objects
- Risks and mitigations
- Validation/test strategy
- Open questions (if any)

Do not include implementation code.

## Completion Criteria

Only finish when:
1. The user has approved the design.
2. The user has chosen one of the allowed save destinations for the doc (`docs/brainstorming/`, target object directory when applicable, custom directory, or skip saving).
3. The brainstorming doc has been written to file (default/target-object/custom) or output inline (skip saving), and shared.
