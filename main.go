package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"kanban-ui/internal/cli"
	"kanban-ui/internal/server"
)

//go:embed ui/dist
var uiFiles embed.FS

func main() {
	if len(os.Args) > 1 && cli.Commands[os.Args[1]] {
		os.Exit(cli.Run(os.Args[1:]))
	}

	kanbanDir := os.Getenv("KANBAN_DIR")
	if kanbanDir == "" {
		kanbanDir = "./kanban"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Generate or load auth token
	token := os.Getenv("AUTH_TOKEN")
	if token == "" {
		token = server.GenerateAuthToken()
	}
	server.SetAuthToken(token)

	log.Printf("Kanban UI starting on http://localhost:%s?token=%s", port, token)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/projects", server.AuthMiddleware(server.HandleAPIListProjects(kanbanDir)))
	mux.HandleFunc("GET /api/projects/{name}", server.AuthMiddleware(server.HandleAPIGetProject(kanbanDir)))
	mux.HandleFunc("GET /api/projects/{project}/columns", server.AuthMiddleware(server.HandleAPIListColumns(kanbanDir)))
	mux.HandleFunc("GET /api/projects/{project}/tickets/{slug}", server.AuthMiddleware(server.HandleAPIGetTicket(kanbanDir)))
	mux.HandleFunc("POST /api/tickets", server.AuthMiddleware(server.HandleAPICreateTicket(kanbanDir)))
	mux.HandleFunc("POST /api/tickets/{slug}/move", server.AuthMiddleware(server.HandleAPIMoveTicket(kanbanDir)))
	mux.HandleFunc("POST /api/tickets/{slug}/field", server.AuthMiddleware(server.HandleAPIUpdateField(kanbanDir)))
	mux.HandleFunc("DELETE /api/tickets/{slug}", server.AuthMiddleware(server.HandleAPIArchiveTicket(kanbanDir)))
	mux.HandleFunc("POST /api/tickets/{slug}/run", server.AuthMiddleware(server.HandleAPIScriptRun(kanbanDir)))
	mux.HandleFunc("GET /api/projects/{project}/config", server.AuthMiddleware(server.HandleAPIGetProjectConfig(kanbanDir)))
	mux.HandleFunc("PUT /api/projects/{project}/config", server.AuthMiddleware(server.HandleAPIUpdateProjectConfig(kanbanDir)))

	go server.StartFileWatcher(kanbanDir)
	mux.HandleFunc("/events", server.HandleSSE)

	distFS, err := fs.Sub(uiFiles, "ui/dist")
	if err != nil {
		log.Fatalf("Failed to create dist sub-FS: %v", err)
	}
	mux.HandleFunc("/", server.SPAHandler(distFS))

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
