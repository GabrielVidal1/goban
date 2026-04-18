# Real-Time UI Refresh via File Watcher + SSE

## Objective

Replace periodic polling with event-driven updates so the Kanban board automatically refreshes when ticket data files change in the kanban directory, without requiring a page reload or unnecessary server requests.

---

## Implementation Plan

### 1. Add `fsnotify` dependency

```bash
go get github.com/fsnotify/fsnotify
```

This Go package watches filesystem paths for changes (create, write, delete, rename).

---

### 2. Create file watcher + SSE broadcaster in `main.go`

Add these functions to `main.go`:

#### File Watcher Setup

```go
var sseClients []chan string // connected SSE clients

func startFileWatcher(kanbanDir string) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Printf("Failed to create file watcher: %v", err)
        return
    }
    defer watcher.Close()

    // Broadcast changes via SSE
    go func() {
        for event := range watcher.Events {
            // Only care about writes, creates, deletes on .md files
            if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 &&
               strings.HasSuffix(event.Name, ".md") {
                log.Printf("File changed: %s", event.Name)

                // Notify all connected SSE clients
                for _, ch := range sseClients {
                    select {
                    case ch <- "refresh":
                    default:
                        close(ch) // remove dead client
                    }
                }
            }
        }
    }()

    // Watch the kanban directory recursively
    err = filepath.Walk(kanbanDir, func(path string, info os.FileInfo, err error) error {
        if err == nil && info.IsDir() {
            watcher.Add(path)
        }
        return nil
    })
    if err != nil {
        log.Printf("Failed to watch directory: %v", err)
    }

    // Block forever so the goroutine stays alive
    select {}
}
```

#### SSE Endpoint Handler

```go
func handleSSE(w http.ResponseWriter, r *http.Request) {
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    ch := make(chan string, 10)
    sseClients = append(sseClients, ch)
    defer func() {
        for i, c := range sseClients {
            if c == ch {
                sseClients = append(sseClients[:i], sseClients[i+1:]...)
                close(ch)
                break
            }
        }
    }()

    flusher.Flush()

    for {
        select {
        case <-r.Context().Done():
            return // client disconnected
        case msg := <-ch:
            fmt.Fprintf(w, "event: filechange\ndata: %s\n\n", msg)
            flusher.Flush()
        }
    }
}
```

#### Wire Up in `main()`

After setting up the mux but before `ListenAndServe`:

```go
// Start file watcher (non-blocking goroutine)
go startFileWatcher(kanbanDir)

// SSE endpoint for real-time updates
mux.HandleFunc("/events", handleSSE)
```

---

### 3. Add HTMX SSE integration to templates

#### In `base.html` — add SSE connection element

Place this right after the opening `<body>` tag, before the header:

```html
<!-- Real-time file change events -->
<div hx-ext="sse" sse-connect="/events" sse-swap="filechange">
</div>
```

This establishes a persistent SSE connection to `/events` and listens for `filechange` events.

#### In `board.html` — trigger refresh on column bodies

Modify the existing `column-body` divs to listen for a custom `refresh` event:

**Before:**
```html
<div class="column-body" hx-get="/board/columns/{{$.CurrentProject}}/{{$col.Name}}/tickets" hx-trigger="revealed">
```

**After:**
```html
<div class="column-body" hx-get="/board/columns/{{$.CurrentProject}}/{{$col.Name}}/tickets" hx-trigger="revealed, refresh from:body">
```

The `refresh from:body` part means it listens for the `refresh` event triggered anywhere on the document body.

#### In `base.html` — add JS listener to bridge SSE → HTMX trigger

Add this to the existing `<script>` block in `base.html`:

```javascript
// Auto-refresh board when kanban files change via SSE
document.body.addEventListener('htmx:sseMessage', function(evt) {
    if (evt.detail.event === 'filechange') {
        // Trigger HTMX reload of all column bodies
        document.querySelectorAll('.column-body').forEach(function(col) {
            htmx.trigger(col, 'refresh');
        });
    }
});
```

---

## Data Flow

```
File change in kanban dir (fsnotify detects .md write/create/delete)
    ↓
Go goroutine broadcasts "refresh" to all SSE client channels
    ↓
SSE endpoint streams event: filechange data: refresh to connected browsers
    ↓
HTMX sse-swap="filechange" fires htmx:sseMessage event in browser
    ↓
JS listener catches the event and triggers 'refresh' on .column-body elements
    ↓
HTMX fetches fresh ticket data from existing /board/columns/.../tickets endpoint
    ↓
UI updates automatically — no page reload needed
```

---

## Trade-offs vs Polling

| Aspect | Polling (`every 3s`) | File Watcher + SSE |
|---|---|---|
| **Server load** | Constant requests every 3s regardless of changes | Zero until a change occurs |
| **Latency** | Up to 3s delay between change and refresh | Near-instant (<100ms) |
| **Complexity** | One-line template change | Requires fsnotify + SSE endpoint + JS bridge |
| **Reliability** | Always works, no persistent connection needed | Depends on long-lived HTTP connection (handled by browser auto-reconnect) |

---

## Verification Criteria

- [ ] `go get github.com/fsnotify/fsnotify` completes successfully and builds pass
- [ ] Starting the server logs "File changed: ..." when a `.md` file in the kanban directory is modified
- [ ] Opening the board page establishes an SSE connection to `/events` (visible in browser DevTools Network tab)
- [ ] Modifying any ticket file in the kanban directory triggers an automatic UI refresh within ~100ms without a full page reload
- [ ] Column ticket counts and card contents update correctly after each change
- [ ] Disconnecting/reconnecting browsers does not leak SSE client channels (dead clients are cleaned up)

---

## Potential Risks and Mitigations

1. **SSE connection drops** — Browsers auto-reconnect SSE connections with exponential backoff by default, so this is handled natively.
2. **Dead client channel leaks** — The `default` case in the broadcast loop closes channels that can't accept a message (buffer full or disconnected). The defer block removes closed channels from the slice.
3. **Watching too many directories** — `filepath.Walk` watches every subdirectory. If the kanban directory has deep nesting with thousands of files, this could be heavy. Mitigation: only watch top-level project dirs and their ticket `.md` files.
4. **Non-.md file changes ignored** — The watcher filters for `.md` suffixes since that's the ticket format. If a different extension is used in the future, update the `strings.HasSuffix` check.

---

## Alternative Approaches

1. **HTMX polling (`every 3s`)** — Simpler (one-line change to `board.html`), but generates constant traffic even when nothing changes. Good for quick prototyping; SSE is better for production.
2. **WebSocket instead of SSE** — Full duplex communication. Overkill for this use case since the server only pushes one-way events to clients. SSE is simpler and has built-in auto-reconnect.
3. **Inotify on Linux / FSEvents on macOS** — Platform-specific watchers. `fsnotify` abstracts these into a cross-platform API, so no need to handle platform differences manually.
