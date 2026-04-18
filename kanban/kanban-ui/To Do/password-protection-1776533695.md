---
title: Password protection
created: 2026-04-18
priority: high
tags: [security, auth]
---

# App Token Authentication — Implementation Plan

## Problem Statement

The Kanban UI application has **zero authentication**. Anyone with network access can read, create, move, and delete tickets. There is no concept of users, sessions, or authorization — all API endpoints are fully public.

## Goals

- Add a single shared app token for authentication.
- Protect all API endpoints behind the token.
- Configure the token via environment variable (`KANBAN_TOKEN`).
- When unset, the app runs in unauthenticated mode (backward compatible).

## Non-Goals

- Password-based login UI.
- Multi-user / role-based access control.
- Token rotation or expiration.
- Per-project tokens.

---

## Architecture Overview

```
┌──────────────┐     HTTPS      ┌─────────────────────────┐
│   Browser    │ ◄────────────► │       Go Server          │
│              │                │                          │
│ Board UI     │  Bearer token  │ authMiddleware()         │
│ (React SPA)  │ ─────────────► │ Check KANBAN_TOKEN       │
│              │                │ 401 if missing/mismatch  │
└──────────────┘                └─────────────────────────┘
```

No login page. No sessions. The token is a shared secret sent as an `Authorization: Bearer <token>` header on every API request.

---

## Backend Changes (Go)

### Configuration

New environment variable:

| Variable | Default | Description |
|---|---|---|
| `KANBAN_TOKEN` | *(empty = disabled)* | The shared token that clients must present to access the API |

When `KANBAN_TOKEN` is set, authentication is **enabled**. When empty/unset, the app runs in unauthenticated mode (backward compatible).

### Middleware

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := os.Getenv("KANBAN_TOKEN")
        if token == "" {
            next.ServeHTTP(w, r) // unauthenticated mode
            return
        }

        authHeader := r.Header.Get("Authorization")
        haveToken := ""
        if strings.HasPrefix(authHeader, "Bearer ") {
            haveToken = authHeader[len("Bearer "):]
        }

        if haveToken == "" || haveToken != token {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

### Route Changes (main.go)

Wrap all API routes with the middleware:

| Endpoint | Wrapped? |
|---|---|
| `GET /api/projects` | Yes |
| `GET /api/projects/{name}` | Yes |
| `POST /api/tickets` | Yes |
| `POST /api/tickets/{slug}/move` | Yes |
| `POST /api/tickets/{slug}/field` | Yes |
| `DELETE /api/tickets/{slug}` | Yes |
| `GET /events` (SSE) | Yes |
| `/` (SPA static files) | **No** — serves the SPA bundle regardless |

### File Changes (Backend)

| File | Change |
|---|---|
| `main.go` | Add `authMiddleware`, wrap API routes, apply to SSE endpoint |

No new dependencies needed. Pure standard library.

---

## Frontend Changes (React/TypeScript)

### API Client (`ui/src/api/client.ts`)

Add the token header to every request:

```ts
const TOKEN = import.meta.env.VITE_KANBAN_TOKEN ?? ''

export async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...options?.headers,
  }
  if (TOKEN) {
    headers['Authorization'] = `Bearer ${TOKEN}`
  }

  const res = await fetch(path, {
    headers,
    ...options,
  })

  if (!res.ok) {
    let message = `HTTP ${res.status}`
    try {
      const body = await res.json()
      if (body.error) message = body.error
    } catch {}
    throw new ApiError(res.status, message)
  }

  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}
```

### Build Config (`ui/vite.config.ts`)

Wire the env var into the build:

```ts
export default defineConfig({
  define: {
    'import.meta.env.VITE_KANBAN_TOKEN': JSON.stringify(
      process.env.VITE_KANBAN_TOKEN ?? ''
    ),
  },
})
```

### Error Handling

On receiving a `401`, show a toast notification: *"Authentication required — check your KANBAN_TOKEN configuration."* No redirect, no login page.

### File Changes (Frontend)

| File | Change |
|---|---|
| `ui/src/api/client.ts` | Add `Authorization: Bearer` header from env var; add 401 toast handling |
| `ui/vite.config.ts` | Wire `VITE_KANBAN_TOKEN` into the build |

No new components needed. No routing changes.

---

## Security Considerations

| Concern | Mitigation |
|---|---|
| Token in transit | Document that HTTPS should be used in production |
| Token exposure in browser DevTools | Acceptable — shared secret, not user-specific |
| Token in logs / history | Use `Bearer` prefix (not Basic) to avoid accidental base64 encoding; document not logging Authorization headers |
| Brute-force | Not applicable — no login endpoint to hammer |

---

## Implementation Steps

### Step 1 — Backend (~15 min)

1. Add `authMiddleware` in `main.go` that checks `KANBAN_TOKEN` env var against the `Authorization: Bearer <token>` header
2. Wrap all API routes and SSE endpoint with the middleware
3. Leave SPA static file serving (`/`) unwrapped so the bundle is always accessible

### Step 2 — Frontend (~15 min)

4. Update `api/client.ts` to read `VITE_KANBAN_TOKEN` from env and attach it as a Bearer header on every request
5. Add 401 error toast in the client or toast context
6. Wire the env var through `vite.config.ts`

### Step 3 — Test (~15 min)

7. Without `KANBAN_TOKEN`: app works exactly as before (no auth)
8. With `KANBAN_TOKEN` set on server but not sent by client: all API calls return 401, SPA loads but shows errors
9. With matching token on both sides: full functionality restored
10. Verify the token is never logged or echoed in responses

---

## Verification Criteria

- [ ] When `KANBAN_TOKEN` is unset, the app behaves exactly as it does today (no auth)
- [ ] When `KANBAN_TOKEN` is set, requests without a matching `Authorization: Bearer` header return `401 Unauthorized`
- [ ] Requests with the correct token succeed normally across all endpoints (API + SSE)
- [ ] The SPA bundle at `/` loads regardless of auth status
- [ ] The token is never logged or returned in any API response body
- [ ] Frontend shows a clear error toast when the server returns 401

---

## Usage

```bash
# Start with authentication enabled:
KANBAN_TOKEN=my-secret-token ./kanban-ui

# Frontend build (token baked into the bundle):
VITE_KANBAN_TOKEN=my-secret-token npm run build
```

Both server and frontend need the same token value for communication to work.
