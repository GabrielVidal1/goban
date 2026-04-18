---
name: kanban
description: >
  Manage a file-based kanban board at kanban/ in the current working directory,
  supporting multiple projects as subdirectories. Each project contains column
  subdirectories, and each column contains ticket .md files with YAML front matter.
  Use this skill whenever the user mentions: kanban, tickets, tasks on a board,
  moving a ticket, creating a ticket, viewing the board, columns, backlog,
  in-progress, done, sprint planning, projects, or any task/project management
  workflow involving their local kanban folder. Always use this skill even for
  simple requests like "show me the board", "list my projects", or "create a ticket".
---

# Kanban Skill

Manages a file-based kanban board at `./kanban/` with multiple project support.

## Board Structure

```
kanban/
  projectA/              ← project (subdirectory of kanban/)
    Backlog/             ← column (subdirectory of project)
      script.sh          ← optional ticket processing script for this column
      my-ticket.md       ← ticket (.md file with YAML front matter)
    In Progress/
      script.sh          ← optional, may differ from Backlog's script
    Done/
    _archive/            ← archived tickets (ignored by default)
  projectB/
    Backlog/
    ...
```

Each column directory may contain a `script.sh` that defines how tickets in that column are processed. For example, a Backlog column's script might run the muse agent to plan a ticket and then move it to To Do.

## Ticket Format

```markdown
---
title: Fix login bug
priority: high        # optional
assignee: alice       # optional
due: 2026-05-01       # optional
tags: [bug, auth]     # optional
created: 2026-04-18   # set automatically on creation
---

Description and notes go here.
```

---

## Scripts

### Column-Level Processing Scripts

Each column directory may contain a `script.sh` that defines how tickets in that column are processed.
Run it with the ticket slug as argument:
```bash
bash kanban/<project>/<column>/script.sh <ticket-slug>
```
For example, to process a Backlog ticket using its column script:
```bash
bash kanban/kanban-ui/Backlog/script.sh split-main-go-into-multiple-files
```

The script receives the ticket slug as `$1`, reads the ticket's front matter and body,
and can perform any processing (e.g., running an agent, generating artifacts) before
moving the ticket to another column.

### Board Utility Scripts

Utility scripts live in `scripts/` alongside this SKILL.md. Run as:
```bash
bash scripts/<script-name> [options]
```
They respect the `KANBAN_DIR` env var (default: `./kanban`).

All utility scripts accept `--project <name>`. If omitted and only one project exists,
it is used automatically. With multiple projects and no `--project`, an error is shown.

---

### `list-projects`
Lists all projects with column and ticket counts.
```bash
bash scripts/list-projects
```

### `list-columns [--project <n>] [--all]`
Lists columns and ticket counts for a project.
`--all` shows columns across every project.
```bash
bash scripts/list-columns --project projectA
bash scripts/list-columns --all
```

### `list-tickets [--project <n>] [--all] [--column <c>] [--assignee <a>] [--priority <p>] [--tag <t>]`
Lists tickets with optional filters.
```bash
bash scripts/list-tickets --project projectA
bash scripts/list-tickets --all
bash scripts/list-tickets --project projectA --column "In Progress"
bash scripts/list-tickets --all --assignee alice --priority high
```

### `get-ticket <slug> [--project <n>]`
Prints the full content of a ticket (searches within the project's columns).
```bash
bash scripts/get-ticket fix-login-bug --project projectA
```

### `new-ticket --column <col> --title <title> [--project <n>] [options]`
Creates a new ticket. Column and project are created if they don't exist.
```bash
bash scripts/new-ticket --project projectA --column Backlog --title "Fix login bug" \
  --priority high --assignee alice --due 2026-05-01 --tags "bug,auth" \
  --body "Users can't log in with SSO."
```

### `update-ticket-status <slug> <target-column> [--project <n>]`
Moves a ticket to a different column.
```bash
bash scripts/update-ticket-status fix-login-bug "In Progress" --project projectA
```

### `update-ticket-field <slug> <field> <value> [--project <n>]`
Updates a single front matter field (inserts it if missing).
```bash
bash scripts/update-ticket-field fix-login-bug priority critical --project projectA
bash scripts/update-ticket-field fix-login-bug assignee bob --project projectA
```

### `archive-ticket <slug> [--project <n>] [--delete] [--yes]`
Archives to `_archive/`, or permanently deletes with `--delete`.
```bash
bash scripts/archive-ticket fix-login-bug --project projectA
bash scripts/archive-ticket fix-login-bug --project projectA --delete --yes
```

---

## Workflows

### Explore the board
```bash
bash scripts/list-projects                        # overview of all projects
bash scripts/list-columns --all                   # all projects' columns
bash scripts/list-tickets --project projectA      # tickets in one project
bash scripts/list-tickets --all                   # every ticket everywhere
```

### Process a ticket from Backlog to To Do
```bash
bash kanban/kanban-ui/Backlog/script.sh split-main-go-into-multiple-files
```
The column's `script.sh` reads the ticket, runs planning (e.g., via muse agent),
and moves it to the next stage.

### Plan a sprint
```bash
bash scripts/list-tickets --project projectA --column Backlog
bash scripts/update-ticket-status my-ticket "In Progress" --project projectA
bash scripts/update-ticket-field my-ticket assignee alice --project projectA
```

### Create and close a ticket
```bash
bash scripts/new-ticket --project projectA --column Backlog --title "New feature"
bash scripts/update-ticket-status new-feature Done --project projectA
bash scripts/archive-ticket new-feature --project projectA
```

### Edit ticket body
Use `str_replace` or `bash_tool` to edit the markdown body below the front matter directly —
the scripts manage front matter fields and file location only.

---

## Notes
- **Auto-project**: if only one project exists, `--project` can be omitted everywhere.
- **Slug matching**: filenames are `<title-slug>-<timestamp>.md`; pass any substring of
  the filename (without `.md`) as the slug — matching is case-insensitive.
- **Column names with spaces**: always quote them: `"In Progress"`.
- **`_archive/`** is excluded from all listing commands by default.
- **Per-column scripts**: each column's `script.sh` is independent — different columns
  can define different processing workflows (e.g., Backlog plans tickets, In Progress
  runs tests).
- Utility scripts have no dependencies beyond standard Unix tools (bash, awk, find, sed).
