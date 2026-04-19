---
name: plan-ticket
description: >
  Plan a kanban ticket in detail: read the ticket, research the codebase, and write
  a step-by-step implementation plan into the ticket body. Once the plan is written,
  move the ticket to the "To Do" column. Use when the user asks to "plan ticket X",
  "design ticket Y", "flesh out a ticket", or "prepare a ticket for implementation".
---

# Plan Ticket

Turn a thin kanban ticket into an implementation-ready plan, then move it to **To Do**.

## Inputs

- **project** (required): kanban project name.
- **slug** (required): ticket slug (substring match is fine).

Ask the user if either is missing. Ensure the binary is built (`make build`) before invoking CLI commands.

## Workflow

### Step 1 — Fetch the ticket

```bash
./kanban-ui ticket get <project> <slug> --json
```

Read the `.md` file at the returned `path` to see the current title, body, and front matter.

### Step 2 — Research

Investigate the codebase enough to write a concrete plan:
- Identify affected files, functions, and modules.
- Note existing patterns to follow (don't invent new abstractions).
- Surface unknowns, risks, and decisions that need user input **before** writing the plan — ask the user to resolve them rather than guessing.

For non-trivial scope, delegate to the `Plan` subagent with a self-contained brief.

### Step 3 — Write the plan into the ticket body

Append (or replace, if a stale plan exists) a `## Plan` section in the ticket `.md` file. Preserve front matter and any existing description above it. Suggested structure:

```markdown
## Plan

### Goal
<1–2 sentences restating the outcome>

### Approach
<short paragraph on the chosen strategy and why>

### Steps
1. <file:line — concrete change>
2. ...

### Files touched
- `path/to/file.ext` — <what changes>

### Acceptance criteria
- [ ] <observable behavior>
- [ ] <test or build passes>

### Risks / open questions
- <risk or follow-up>
```

Keep it concrete: reference real files and line numbers, not hypotheticals. No filler.

### Step 4 — Move to To Do

```bash
./kanban-ui ticket move <project> <slug> "To Do"
```

If the column name differs, inspect with `./kanban-ui columns list <project>` and use the matching one. Confirm with the user if there's no obvious "To Do"-equivalent column.

### Step 5 — Report

Tell the user:
- ticket path
- new column
- a one-line summary of the plan

Do not start implementing — that's `auto-implement`'s job.

## Error Handling

- **Slug ambiguous**: list with `tickets list --json` and retry with the exact slug.
- **Open questions remain**: stop and ask the user before writing the plan; an unresolved plan is worse than no plan.
- **Ticket already has a plan**: ask whether to replace, extend, or abort.
