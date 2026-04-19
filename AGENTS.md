# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

- `make dev` — parallel dev: `air` rebuilds/restarts Go on `.go` changes, `vite` serves the UI on its own port with HMR. The Vite dev server proxies `/api` and `/events` to `localhost:8080` (see `ui/vite.config.ts`), so open the Vite URL during UI work.
- `make build` — builds the UI (`npm run build` → `ui/dist`), then compiles the Go binary with `CGO_ENABLED=0`. The UI is embedded via `//go:embed ui/dist` in `main.go`, so rebuilding the binary is required to ship UI changes in production mode.
- `make run` — builds then runs under `sudo -E` (needed only when binding to a privileged port). `make run-local` runs without sudo.
- `make format` — `gofmt -w .`
- `make tools` — installs `air` as a Go tool (used by `dev-go`).
- `make clean` — removes the binary, `ui/dist`, and `tmp/`.
- Single Go test: no tests exist yet. If adding: `go test ./internal/kanban -run TestName`.

The binary boots in **server mode** by default. If `os.Args[1]` is in `cli.Commands`, it runs in **CLI mode** instead and exits (see `main.go:20-23`). Run `./kanban-ui help` for the full CLI surface.

## Architecture

**Single-binary Go server + embedded React SPA.** The Go process serves both the JSON API (`/api/*`), the SSE stream (`/events`), and the static SPA (everything else, falling back to `index.html` for client-side routes via `server/spa.go`).

### Data model — everything is files on disk

Source of truth is `KANBAN_DIR` (default `./kanban`):

```
kanban/
  <project>/
    config.json              # { "columnsOrder": [...] }
    <column>/
      <slug>-<unix>.md       # YAML-ish front matter + body
      script.sh              # optional, runnable per-ticket
    _archive/                # folders prefixed "_" are hidden
```

- A directory is considered a **project** only if it has ≥1 subdirectory and does not start with `_` (see `kanban.ListProjects`).
- Column ordering: `ApplyColumnOrder` places configured columns first in the given order, then appends unlisted ones alphabetically. Entries in `columnsOrder` that don't exist on disk are silently ignored.
- Ticket front matter is parsed by a bespoke line-splitter in `ParseTicket` — **not** a real YAML parser. Only `title`, `priority`, `assignee`, `due`, `tags`, `created` are recognized. `tags` is a bracketed comma list. Filename (minus `.md`) is the slug.
- **Ticket lookup is substring-match on filename**, not exact slug match (`GetTicket`, `UpdateTicketStatus`, `UpdateTicketField`, `ArchiveTicket`, `RunScript` all use `strings.Contains(nameLower, slugLower)`). This is intentional for UX but means passing a short/ambiguous slug matches the first file walked.
- `UpdateTicketStatus` moves a ticket by `os.Rename` across column directories; this preserves the timestamped filename.
- `script.sh` is executed by `RunScript` with `bash script.sh <slug>`, cwd set to the column folder.

### Config precedence (`internal/config`)

`config.Load(flagDir, flagPort)` applies in order: flag → env var → `.env` file → hardcoded default. `.env` is read manually (no external dep) and only sets vars that aren't already in the environment. Vars: `KANBAN_DIR`, `PORT`, `AUTH_TOKEN`.

### Auth

- On server boot, if `AUTH_TOKEN` is unset, a random 32-byte hex token is generated and logged in the startup URL (`main.go:37`).
- All `/api/*` routes go through `AuthMiddleware`, which accepts the token from `?token=` query param **or** `X-Auth-Token` header.
- The SPA persists the token to `localStorage` on first load (extracted from the `?token=` URL), then sends it as `X-Auth-Token` on every `apiFetch` call. A 401 dispatches a `kanban:unauthorized` window event, which `App.tsx` surfaces as a toast.
- `/events` (SSE) is **not** behind auth middleware.

### Realtime refresh

`StartFileWatcher` (spawned as a goroutine in `main.go`) uses `fsnotify` to recursively add every directory under `KANBAN_DIR` at startup, then broadcasts a `"refresh"` string to every SSE client on any write/create/remove. The UI's `useBoard` hook subscribes to `/events` and calls `fetchBoard()` on each `filechange` event. Note: the watcher only walks directories **at startup** — new subdirs created during runtime won't be watched until restart.

### CLI

`internal/cli/cli.go` is a hand-rolled dispatcher (no cobra/urfave). Top-level commands are gated by the `Commands` map in `cli.go` — adding a new top-level verb requires registering it there or `main.go` will treat it as a server flag. `parseArgs` is a custom flag parser that lets flags appear before or after positional args (Go's `flag` stops at the first non-flag token by default).

### API surface

Routes are declared in `main.go:41-54` using Go 1.22+ pattern syntax (`GET /api/...`, `{slug}` path values). The server reuses `internal/kanban` for all disk operations — the HTTP handlers are thin wrappers around the same functions the CLI calls. `sanitizeTicket` in `server/helpers.go` strips the `path` field before returning tickets to the UI.

### Frontend

- React 19 + Vite 6 + Tailwind 3 + React Router 7 + dnd-kit for drag-and-drop columns/tickets.
- `ui/src/api/client.ts` is the single fetch wrapper — all API calls go through `apiFetch` so auth and 401 handling stay centralized.
- `ui/src/hooks/useBoard.ts` owns board state + the SSE subscription; components dispatch a `kanban:refresh` custom event for imperative refreshes (e.g. after creating a ticket).
- Theme is persisted in `localStorage` under `theme`; auth token under `kanban_auth_token`.
