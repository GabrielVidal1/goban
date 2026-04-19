---
name: auto-implement
description: >
  Automatically implement a kanban ticket end-to-end: fetch the ticket, move it to "In Progress",
  develop the feature using the forge agent, and update the ticket status. Use when the user asks
  you to implement or build a specific ticket from the kanban board (e.g., "implement ticket X",
  "build this feature", "work on ticket Y"). Triggers on any request to auto-implement a kanban
  ticket rather than doing it manually step by step.
---

# Auto-Implement

End-to-end ticket implementation: fetch → move to In Progress → develop → mark Done.

## Prerequisites

Ensure the kanban binary is built before running any commands:

```bash
make build
```

If `KANBAN_DIR` or `--dir` is needed, set it in every command invocation.

## Workflow

### Step 1 — Identify the ticket

Ask the user which project and ticket to implement if not already specified. Resolve the slug using substring matching (case-insensitive).

```bash
./kanban-ui tickets list <project> --json
./kanban-ui ticket get <project> <slug> --json
```

### Step 2 — Move to In Progress

Move the ticket to the "In Progress" column before starting work.

```bash
./kanban-ui ticket move <project> <slug> "In Progress"
```

If the target column name differs (e.g., "To Do"), use `ticket get --json` to inspect current columns and adapt.

### Step 3 — Read the ticket body

Fetch the full ticket details including the file path:

```bash
./kanban-ui ticket get <project> <slug> --json
```

Read the `.md` file at the returned `path` to understand requirements, acceptance criteria, and any implementation notes.

### Step 5 — Validate and mark Done

Once the forge agent completes:

1. Verify the changes compile/build successfully.
2. Run any relevant tests.
3. Update the ticket body with implementation notes if not already done by forge.
4. Move the ticket to "To Validate":

```bash
./kanban-ui ticket move <project> <slug> To Validate
```

Optionally archive:

```bash
./kanban-ui ticket archive <project> <slug>
```

## Error Handling

- **Column not found**: List columns with `columns list <project>` and use the correct name.
- **Slug ambiguous**: Use `tickets list --json` to find the exact slug, then retry.
- **Build fails after implementation**: Fix compilation errors before marking Done. Report failures back to the user.
