package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"kanban-ui/internal/server"
)

//go:embed ui/dist
var uiFiles embed.FS

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

	mux.HandleFunc("GET /api/projects", server.HandleAPIListProjects(kanbanDir))
	mux.HandleFunc("GET /api/projects/{name}", server.HandleAPIGetProject(kanbanDir))
	mux.HandleFunc("GET /api/projects/{project}/columns", server.HandleAPIListColumns(kanbanDir))
	mux.HandleFunc("GET /api/projects/{project}/tickets/{slug}", server.HandleAPIGetTicket(kanbanDir))
	mux.HandleFunc("POST /api/tickets", server.HandleAPICreateTicket(kanbanDir))
	mux.HandleFunc("POST /api/tickets/{slug}/move", server.HandleAPIMoveTicket(kanbanDir))
	mux.HandleFunc("POST /api/tickets/{slug}/field", server.HandleAPIUpdateField(kanbanDir))
	mux.HandleFunc("DELETE /api/tickets/{slug}", server.HandleAPIArchiveTicket(kanbanDir))
	mux.HandleFunc("POST /api/tickets/{slug}/run", server.HandleAPIScriptRun(kanbanDir))

	go server.StartFileWatcher(kanbanDir)
	mux.HandleFunc("/events", server.HandleSSE)

	distFS, err := fs.Sub(uiFiles, "ui/dist")
	if err != nil {
		log.Fatalf("Failed to create dist sub-FS: %v", err)
	}
	mux.HandleFunc("/", server.SPAHandler(distFS))

	log.Printf("Kanban UI starting on :%s (KANBAN_DIR=%s)", port, kanbanDir)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
