package server

import (
	"encoding/json"
	"net/http"

	"kanban-ui/internal/kanban"
)

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
