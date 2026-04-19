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

Manages a file-based kanban board at `./kanban/` with multiple project support,
driven by the `./kanban-ui` binary's CLI mode.

## Board Structure

```
kanban/
  projectA/              ← project (subdirectory of kanban/)
    Backlog/             ← column (subdirectory of project)
      my-ticket.md       ← ticket (.md file with YAML front matter)
    In Progress/
    Done/
    _archive/            ← archived tickets (ignored by default)
    config.json          ← optional; { "columnsOrder": ["Backlog", ...] }
  projectB/
    Backlog/
    ...
```

## Ticket Format

```markdown
---
title: Fix login bug
priority: high # optional
assignee: alice # optional
due: 2026-05-01 # optional
tags: [bug, auth] # optional
created: 2026-04-18 # set automatically on creation
---

Description and notes go here.
```

---

## CLI

All operations run through the `kanban-ui` binary.
then invoke the subcommands below. Commands respect `KANBAN_DIR` (default `./kanban`);
the `--dir` flag overrides the env var. Add `--json` to any list/get command for
machine-readable output.

```bash
./kanban-ui <command> [subcommand] [args] [flags]
./kanban-ui help          # full usage reference
```

> The CLI always requires a `<project>` argument — no auto-detection. If the user
> doesn't specify one and only one project exists, look it up first with
> `./kanban-ui projects list`.

---

### Projects

```bash
./kanban-ui projects list                   # name, columns, ticket count
./kanban-ui project info <project>
```

### Columns

```bash
./kanban-ui columns list <project>          # in configured order
```

### Tickets

```bash
./kanban-ui tickets list <project> \
    [--column <c>] [--assignee <a>] [--priority <p>] [--tag <t>]

./kanban-ui ticket get <project> <slug>

./kanban-ui ticket create <project> <column> <title> \
    [--priority <p>] [--assignee <a>] [--due <yyyy-mm-dd>] \
    [--tags "a,b,c"] [--body "..."] [--body-file PATH|-]

# Update any combination of fields in ONE call. Only flags you pass are changed.
# Prefer this over multiple `ticket set` invocations.
./kanban-ui ticket edit <project> <slug> \
    [--title <t>] [--priority <p>] [--assignee <a>] [--due <d>] \
    [--tags "a,b,c"] [--body "..."] [--body-file PATH|-]

./kanban-ui ticket move <project> <slug> <target-column>

./kanban-ui ticket set <project> <slug> <field> <value>     # single-field; prefer `ticket edit`

./kanban-ui ticket archive <project> <slug> [--delete]

./kanban-ui ticket run <project> <slug>     # executes column's script.sh
```

### Config

```bash
./kanban-ui config get <project>
./kanban-ui config set-order <project> "Backlog,To Do,In Progress,Done"
```

---

## Workflows

### Explore the board

```bash
./kanban-ui projects list
./kanban-ui columns list projectA
./kanban-ui tickets list projectA
```

To see tickets across every project, iterate over `projects list --json`
and call `tickets list <project>` per project.

### Plan a sprint

```bash
./kanban-ui tickets list projectA --column Backlog
./kanban-ui ticket move projectA fix-login-bug "In Progress"
./kanban-ui ticket set projectA fix-login-bug assignee alice
```

### Create and close a ticket

```bash
./kanban-ui ticket create projectA Backlog "New feature"
./kanban-ui ticket move projectA new-feature Done
./kanban-ui ticket archive projectA new-feature
```

### Edit ticket body

Use `ticket edit --body "..."` (or `--body-file PATH`, or `--body-file -` to
pipe from stdin for long bodies). For in-place edits that preserve the rest of
the file, find the path via `ticket get --json` and use the `Edit` tool.

---

## Notes

- **Slug matching**: filenames are `<title-slug>-<timestamp>.md`; pass any
  substring of the filename (without `.md`) as the slug — matching is case-insensitive.
- **Column names with spaces**: always quote them: `"In Progress"`.
- **`_archive/`** is excluded from all listing commands by default.
- **Exit codes**: 0 on success, 1 on operational error, 2 on usage error.
  Errors go to stderr.
- If `./kanban-ui` isn't built, run `make build` from the project root first.
