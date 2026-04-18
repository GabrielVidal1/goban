---
title: Configuration via .env file or CLI flags
priority: medium
tags: [config,dx]
created: 2026-04-18
---

## Context

`main.go` already reads two config values from environment variables:

| Variable | Default | Where used |
|---|---|---|
| `KANBAN_DIR` | `./kanban` | kanban directory path |
| `PORT` | `8080` | HTTP listen port |

There is no CLI flag support and no `.env` file loading. This ticket adds both, with a clear precedence order.

## Precedence (highest → lowest)

```
CLI flags  >  env vars  >  .env file  >  hardcoded defaults
```

## Config Values

| CLI flag | Env var | Default | Description |
|---|---|---|---|
| `--dir` | `KANBAN_DIR` | `./kanban` | Path to kanban directory |
| `--port` | `PORT` | `8080` | HTTP listen port |

No new config values are needed beyond what already exists.

---

## Implementation Plan

### Step 1 — Add CLI flag parsing (`main.go`)

Use the stdlib `flag` package — no new dependency.

```go
import "flag"

func main() {
    dir  := flag.String("dir",  "", "kanban directory (overrides KANBAN_DIR)")
    port := flag.String("port", "", "listen port (overrides PORT)")
    flag.Parse()
    // resolved below
}
```

### Step 2 — Load `.env` file

Parse a `.env` file manually — no external dependency needed. Only set env vars that are **not already set** (so real env vars keep precedence over the file).

```go
func loadDotEnv(path string) {
    f, err := os.Open(path)
    if err != nil {
        return // .env is optional; silently skip if missing
    }
    defer f.Close()
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        key, val, ok := strings.Cut(line, "=")
        if !ok {
            continue
        }
        key = strings.TrimSpace(key)
        val = strings.Trim(strings.TrimSpace(val), `"'`)
        if os.Getenv(key) == "" { // don't override real env vars
            os.Setenv(key, val)
        }
    }
}
```

Call `loadDotEnv(".env")` **before** reading any `os.Getenv` calls and **before** resolving CLI flags.

### Step 3 — Resolve final config values

```go
func main() {
    dir  := flag.String("dir",  "", "kanban directory (overrides KANBAN_DIR)")
    port := flag.String("port", "", "listen port (overrides PORT)")
    flag.Parse()

    loadDotEnv(".env")

    kanbanDir := *dir
    if kanbanDir == "" {
        kanbanDir = os.Getenv("KANBAN_DIR")
    }
    if kanbanDir == "" {
        kanbanDir = "./kanban"
    }

    listenPort := *port
    if listenPort == "" {
        listenPort = os.Getenv("PORT")
    }
    if listenPort == "" {
        listenPort = "8080"
    }
    // rest of main unchanged
}
```

### Step 4 — `.env.example` file

Add a committed `.env.example` so users know what's configurable:

```
# .env.example — copy to .env and customize
KANBAN_DIR=./kanban
PORT=8080
```

`.env` itself goes in `.gitignore`.

---

## File Changes

| File | Change |
|---|---|
| `main.go` | Add `flag` parsing, `loadDotEnv`, resolve config with precedence |
| `.env.example` | New file — documents available variables |
| `.gitignore` | Add `.env` entry |

No new dependencies. No changes to the frontend or API.

---

## Verification

- `./kanban-ui` → uses defaults (`./kanban`, port `8080`)
- `./kanban-ui --port 9090 --dir /data/kanban` → CLI flags win
- `PORT=3000 ./kanban-ui` → env var used (`3000`)
- `PORT=3000 ./kanban-ui --port 9090` → CLI flag wins (`9090`)
- `.env` with `PORT=4000`, no env var, no flag → `.env` used (`4000`)
- `.env` absent → silently ignored, defaults apply
