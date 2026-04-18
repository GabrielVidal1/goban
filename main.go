package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"

	"kanban-ui/internal/cli"
	"kanban-ui/internal/config"
	"kanban-ui/internal/server"
)

//go:embed ui/dist
var uiFiles embed.FS

func main() {
	// CLI mode: load config from env/.env, then dispatch.
	if len(os.Args) > 1 && cli.Commands[os.Args[1]] {
		os.Exit(cli.Run(os.Args[1:], config.Load("", "")))
	}

	// Server mode: parse flags first so they can override env/.env.
	dir := flag.String("dir", "", "kanban directory (overrides KANBAN_DIR)")
	port := flag.String("port", "", "listen port (overrides PORT)")
	flag.Parse()

	cfg := config.Load(*dir, *port)

	if cfg.AuthToken == "" {
		cfg.AuthToken = server.GenerateAuthToken()
	}
	server.SetAuthToken(cfg.AuthToken)

	log.Printf("Kanban UI starting on http://localhost:%s?token=%s", cfg.Port, cfg.AuthToken)

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

	distFS, err := fs.Sub(uiFiles, "ui/dist")
	if err != nil {
		log.Fatalf("Failed to create dist sub-FS: %v", err)
	}
	mux.HandleFunc("/", server.SPAHandler(distFS))

	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
