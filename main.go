package main

import (
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"kanban-ui/internal/kanban"
)

var (
	sseClients   []chan string
	sseClientsMu sync.Mutex
)

//go:embed ui/dist
var uiFiles embed.FS

// ticketJSON suppresses the internal filesystem path from API responses.
type ticketJSON struct {
	kanban.Ticket
	Path string `json:"path,omitempty"`
}

func main() {
	kanbanDir := os.Getenv("KANBAN_DIR")
	if kanbanDir == "" {
		kanbanDir = "./kanban"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// JSON API
	mux.HandleFunc("GET /api/projects", handleAPIListProjects(kanbanDir))
	mux.HandleFunc("GET /api/projects/{name}", handleAPIGetProject(kanbanDir))
	mux.HandleFunc("GET /api/projects/{project}/columns", handleAPIListColumns(kanbanDir))
	mux.HandleFunc("GET /api/projects/{project}/tickets/{slug}", handleAPIGetTicket(kanbanDir))
	mux.HandleFunc("POST /api/tickets", handleAPICreateTicket(kanbanDir))
	mux.HandleFunc("POST /api/tickets/{slug}/move", handleAPIMoveTicket(kanbanDir))
	mux.HandleFunc("POST /api/tickets/{slug}/field", handleAPIUpdateField(kanbanDir))
	mux.HandleFunc("DELETE /api/tickets/{slug}", handleAPIArchiveTicket(kanbanDir))

	// SSE
	go startFileWatcher(kanbanDir)
	mux.HandleFunc("/events", handleSSE)

	// SPA
	distFS, err := fs.Sub(uiFiles, "ui/dist")
	if err != nil {
		log.Fatalf("Failed to create dist sub-FS: %v", err)
	}
	mux.HandleFunc("/", spaHandler(distFS))

	log.Printf("Kanban UI starting on :%s (KANBAN_DIR=%s)", port, kanbanDir)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// ---- SPA Handler ----

func spaHandler(distFS fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(distFS))
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		f, err := distFS.Open(path)
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		// Fall through to index.html for React Router client-side routes
		index, err := distFS.Open("index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer index.Close()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.Copy(w, index)
	}
}

// ---- JSON helpers ----

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func sanitizeTicket(t *kanban.Ticket) map[string]any {
	return map[string]any{
		"title":    t.Title,
		"priority": t.Priority,
		"assignee": t.Assignee,
		"due":      t.Due,
		"tags":     t.Tags,
		"created":  t.Created,
		"body":     t.Body,
		"slug":     t.Slug,
		"column":   t.Column,
		"project":  t.Project,
	}
}

// ---- API Handlers ----

func handleAPIListProjects(kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projects, err := kanban.ListProjects(kanbanDir)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		type projectSummary struct {
			Name        string   `json:"name"`
			Columns     []string `json:"columns"`
			TicketCount int      `json:"ticket_count"`
		}

		result := make([]projectSummary, 0, len(projects))
		for _, p := range projects {
			info, err := kanban.GetProjectInfo(kanbanDir, p)
			if err != nil {
				continue
			}
			result = append(result, projectSummary{
				Name:        info.Name,
				Columns:     info.Columns,
				TicketCount: info.TicketCount,
			})
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func handleAPIGetProject(kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		board, err := kanban.GetBoardData(kanbanDir, name)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}

		type colJSON struct {
			Name    string           `json:"name"`
			Tickets []map[string]any `json:"tickets"`
		}
		type boardJSON struct {
			Project string    `json:"project"`
			Columns []colJSON `json:"columns"`
		}

		cols := make([]colJSON, len(board.Columns))
		for i, col := range board.Columns {
			tickets := make([]map[string]any, len(col.Tickets))
			for j := range col.Tickets {
				tickets[j] = sanitizeTicket(&col.Tickets[j])
			}
			cols[i] = colJSON{Name: col.Name, Tickets: tickets}
		}
		writeJSON(w, http.StatusOK, boardJSON{Project: board.Project, Columns: cols})
	}
}

func handleAPIListColumns(kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		project := r.PathValue("project")
		cols, err := kanban.ListColumns(kanbanDir, project)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, cols)
	}
}

func handleAPIGetTicket(kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		project := r.PathValue("project")
		slug := r.PathValue("slug")
		ticket, err := kanban.GetTicket(kanbanDir, project, slug)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, sanitizeTicket(ticket))
	}
}

func handleAPICreateTicket(kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Project  string `json:"project"`
			Column   string `json:"column"`
			Title    string `json:"title"`
			Priority string `json:"priority"`
			Assignee string `json:"assignee"`
			Due      string `json:"due"`
			Tags     string `json:"tags"`
			Body     string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.Title == "" {
			writeError(w, http.StatusBadRequest, "title is required")
			return
		}
		if req.Column == "" {
			writeError(w, http.StatusBadRequest, "column is required")
			return
		}

		opts := map[string]string{}
		if req.Priority != "" {
			opts["priority"] = req.Priority
		}
		if req.Assignee != "" {
			opts["assignee"] = req.Assignee
		}
		if req.Due != "" {
			opts["due"] = req.Due
		}
		if req.Tags != "" {
			opts["tags"] = req.Tags
		}
		if req.Body != "" {
			opts["body"] = req.Body
		}

		ticket, err := kanban.CreateTicket(kanbanDir, req.Project, req.Column, req.Title, opts)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, sanitizeTicket(ticket))
	}
}

func handleAPIMoveTicket(kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		var req struct {
			Project      string `json:"project"`
			TargetColumn string `json:"target_column"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.Project == "" || req.TargetColumn == "" {
			writeError(w, http.StatusBadRequest, "project and target_column are required")
			return
		}

		ticket, err := kanban.UpdateTicketStatus(kanbanDir, req.Project, slug, req.TargetColumn)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, sanitizeTicket(ticket))
	}
}

func handleAPIUpdateField(kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		var req struct {
			Project string `json:"project"`
			Field   string `json:"field"`
			Value   string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.Field == "" || req.Project == "" {
			writeError(w, http.StatusBadRequest, "project and field are required")
			return
		}

		ticket, err := kanban.UpdateTicketField(kanbanDir, req.Project, slug, req.Field, req.Value)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, sanitizeTicket(ticket))
	}
}

func handleAPIArchiveTicket(kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		project := r.URL.Query().Get("project")
		if project == "" {
			writeError(w, http.StatusBadRequest, "project query parameter is required")
			return
		}

		err := kanban.ArchiveTicket(kanbanDir, project, slug, false)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "archived"})
	}
}

// ---- SSE + File Watcher ----

func startFileWatcher(kanbanDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create file watcher: %v", err)
		return
	}
	defer watcher.Close()

	go func() {
		for event := range watcher.Events {
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 &&
				strings.HasSuffix(event.Name, ".md") {
				log.Printf("File changed: %s", event.Name)
				sseClientsMu.Lock()
				alive := sseClients[:0]
				for _, ch := range sseClients {
					select {
					case ch <- "refresh":
						alive = append(alive, ch)
					default:
						close(ch)
					}
				}
				sseClients = alive
				sseClientsMu.Unlock()
			}
		}
	}()

	filepath.Walk(kanbanDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			watcher.Add(path)
		}
		return nil
	})

	select {}
}

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
	sseClientsMu.Lock()
	sseClients = append(sseClients, ch)
	sseClientsMu.Unlock()

	defer func() {
		sseClientsMu.Lock()
		for i, c := range sseClients {
			if c == ch {
				sseClients = append(sseClients[:i], sseClients[i+1:]...)
				close(ch)
				break
			}
		}
		sseClientsMu.Unlock()
	}()

	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-ch:
			w.Write([]byte("event: filechange\ndata: " + msg + "\n\n"))
			flusher.Flush()
		}
	}
}

// ---- Helpers ----

func getProjectList(kanbanDir string) []string {
	projects, err := kanban.ListProjects(kanbanDir)
	if err != nil {
		return []string{}
	}
	return projects
}
