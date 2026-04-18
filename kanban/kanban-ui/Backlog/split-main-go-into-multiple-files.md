---
title: Split main.go into multiple files
priority: medium
tags: [refactoring, architecture]
created: 2026-04-18
---

# Split main.go into multiple files

## Problem
`main.go` (449 lines) is a single monolithic file that mixes several concerns:
- Server setup and routing configuration
- API handlers for all endpoints (list projects, get project, list columns, CRUD tickets, script execution)
- SPA static file handler
- JSON response helpers
- SSE + File Watcher logic

This makes the codebase hard to navigate, test, and maintain.

## Proposed Solution
Split `main.go` into a small directory of focused files:

```
cmd/server/
  main.go          # Entry point: wiring everything together, env config, server startup
  handlers.go      # HTTP handler functions (API + SPA)
  api_projects.go  # Project-related API handlers
  api_tickets.go   # Ticket CRUD API handlers
  spa_handler.go   # Static file / SPA serving logic
  sse.go           # Server-Sent Events handler
  watcher.go       # File watcher goroutine
  helpers.go       # JSON response helpers, sanitization utilities
```

Each file should have a single responsibility and be independently testable.

## Acceptance Criteria
- [ ] `main.go` is reduced to ~50 lines (entry point only)
- [ ] All handlers are extracted into separate files organized by concern
- [ ] No functional changes — behavior must remain identical
- [ ] Existing tests still pass
- [ ] Build succeeds without errors

## Notes
- The `kanban/` package already handles domain logic, so the split is purely about HTTP layer organization.
- Consider keeping `helpers.go` for shared utilities like `writeJSON`, `writeError`, and `sanitizeTicket`.
- SSE and file watcher are tightly coupled — they can live in one file or be split if needed.
