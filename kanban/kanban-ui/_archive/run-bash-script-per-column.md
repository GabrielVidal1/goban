---
title: Run bash script per column from stdin (./script.sh <ticket>)
priority: medium
tags: [feature, automation]
created: 2026-04-18
---

## Problem

Currently there is no way to run a custom bash script on demand for a specific ticket.

## Requirement

Implement support for running `./script.sh <ticket>` where the script content is piped via **stdin** and the column folder is resolved from the ticket's location.

### Behavior

1. User runs: `cat column/script.sh | ./script.sh <ticket-slug> --project <project-name>`
2. The system resolves which column the ticket lives in (e.g., `kanban/kanban-ui/Backlog/`).
3. It reads the script from **stdin**.
4. It executes the script, passing the ticket slug as an argument.
5. If no matching ticket is found, print a clear error message.

### Example

```bash
# Ticket lives in kanban/kanban-ui/Backlog/my-ticket.md
# Script content piped via stdin

cat kanban/kanban-ui/Backlog/script.sh | ./script.sh my-ticket --project kanban-ui
# → runs the script with "my-ticket" as argument
```

### Acceptance Criteria

- [ ] Resolves ticket location by slug within the specified project.
- [ ] Reads script content from stdin.
- [ ] Passes ticket slug to the script as an argument.
- [ ] Clear error when no matching ticket is found.
- [ ] Works with columns that have spaces in names (e.g., `"In Progress"`).
