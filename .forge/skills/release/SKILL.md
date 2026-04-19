---
name: release
description: >
  Release a kanban project: move all tickets from the Done column into
  kanban/<project>/_releases/<version>/ and write a RELEASE.md note summarizing
  what shipped. Use when the user asks to "release", "cut a release", "ship",
  or "tag a version" for a specific kanban project.
---

# Release

Snapshot the Done column into a versioned release folder and generate release notes.

## Inputs

- **project** (required): kanban project name (subdir of `KANBAN_DIR`, default `./kanban`).
- **version** (required): release identifier, e.g. `v0.3.0` or `2026.04.19`. Ask the user if not provided. Suggest the next semver based on existing folders under `_releases/`.

## Workflow

### Step 1 — Resolve paths

Determine `KANBAN_DIR` (env var or `./kanban`). Project root is `<KANBAN_DIR>/<project>`. Release dir is `<project_root>/_releases/<version>/`.

Refuse if the release dir already exists — ask the user to pick a new version or confirm overwrite.

### Step 2 — Collect Done tickets

List `.md` files in the Done column directory (typical name `Done`; if missing, inspect with `./kanban-ui columns list <project>` and pick the matching column). Read each ticket's front matter (title, priority, assignee, tags, created) and body for the release note.

If Done is empty, stop and tell the user there's nothing to release.

### Step 3 — Move tickets

Create `<project_root>/_releases/<version>/` and move each Done `.md` file into it (preserve filenames). Use `git mv` if the project is in a git repo, otherwise `mv`. Do not move `script.sh` or other non-ticket files.

Note: folders prefixed with `_` are hidden from the board (see `kanban.ListProjects`), so `_releases/` will not appear as a project — this is intentional.

### Step 4 — Write RELEASE.md

Create `<project_root>/_releases/<version>/RELEASE.md` with this structure:

```markdown
# Release <version>

_Released on <YYYY-MM-DD>._

## Summary

<1–3 sentence overview synthesized from the tickets>

## Tickets (<count>)

- **<title>** (`<slug>`) — <one-line summary from body or first paragraph>
  - priority: <priority> · assignee: <assignee> · tags: <tags>
```

Group by tag or priority if it improves readability. Keep it concise — this is a changelog, not a design doc.

### Step 5 — Report

Tell the user:
- release path
- number of tickets moved
- path to `RELEASE.md`

Do not commit or push automatically — leave that to the user.

## Error Handling

- **Project not found**: list projects with `./kanban-ui projects list` and ask.
- **Done column missing**: inspect columns and confirm the right one with the user.
- **Release dir exists**: stop and ask before overwriting.
