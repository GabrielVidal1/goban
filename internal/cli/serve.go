package cli

import (
	"io/fs"
	"log"
	"net/http"

	"kanban-ui/internal/server"
)

// UIFiles must be set by main before Run is called. It provides the embedded
// ui/dist filesystem served by the "serve" command.
var UIFiles fs.FS

func cmdServe(_ []string) int {
	cfg := globalConfig

	if cfg.AuthToken == "" {
		cfg.AuthToken = server.GenerateAuthToken()
	}
	server.SetAuthToken(cfg.AuthToken)

	log.Printf("Kanban UI starting on http://%s:%s", cfg.Host, cfg.Port)
	log.Printf("Auth token: %s", cfg.AuthToken)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/projects", server.AuthMiddleware(server.HandleAPIListProjects(cfg.KanbanDir)))
	mux.HandleFunc("GET /api/projects/{name}", server.AuthMiddleware(server.HandleAPIGetProject(cfg.KanbanDir)))
	mux.HandleFunc("GET /api/projects/{project}/columns", server.AuthMiddleware(server.HandleAPIListColumns(cfg.KanbanDir)))
	mux.HandleFunc("GET /api/projects/{project}/tickets/{slug}", server.AuthMiddleware(server.HandleAPIGetTicket(cfg.KanbanDir)))
	mux.HandleFunc("POST /api/tickets", server.AuthMiddleware(server.HandleAPICreateTicket(cfg.KanbanDir)))
	mux.HandleFunc("POST /api/tickets/{slug}/move", server.AuthMiddleware(server.HandleAPIMoveTicket(cfg.KanbanDir)))
	mux.HandleFunc("POST /api/tickets/{slug}/field", server.AuthMiddleware(server.HandleAPIUpdateField(cfg.KanbanDir)))
	mux.HandleFunc("DELETE /api/tickets/{slug}", server.AuthMiddleware(server.HandleAPIArchiveTicket(cfg.KanbanDir)))
	mux.HandleFunc("POST /api/tickets/{slug}/run", server.AuthMiddleware(server.HandleAPIScriptRun(cfg.KanbanDir)))
	mux.HandleFunc("GET /api/projects/{project}/config", server.AuthMiddleware(server.HandleAPIGetProjectConfig(cfg.KanbanDir)))
	mux.HandleFunc("PUT /api/projects/{project}/config", server.AuthMiddleware(server.HandleAPIUpdateProjectConfig(cfg.KanbanDir)))

	go server.StartFileWatcher(cfg.KanbanDir)
	mux.HandleFunc("/events", server.HandleSSE)

	// Health and readiness probes (no auth required).
	mux.HandleFunc("GET /health", server.HealthHandler)
	mux.HandleFunc("GET /ready", server.ReadyHandler)

	if UIFiles != nil {
		mux.HandleFunc("/", server.SPAHandler(UIFiles))
	}

	if err := http.ListenAndServe(cfg.Host+":"+cfg.Port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
	return 0
}
