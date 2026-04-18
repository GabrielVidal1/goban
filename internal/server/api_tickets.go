package server

import (
	"encoding/json"
	"net/http"

	"kanban-ui/internal/kanban"
)

func HandleAPICreateTicket(kanbanDir string) http.HandlerFunc {
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

func HandleAPIMoveTicket(kanbanDir string) http.HandlerFunc {
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

func HandleAPIUpdateField(kanbanDir string) http.HandlerFunc {
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

func HandleAPIArchiveTicket(kanbanDir string) http.HandlerFunc {
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

func HandleAPIScriptRun(kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		var req struct {
			Project string `json:"project"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.Project == "" {
			writeError(w, http.StatusBadRequest, "project is required")
			return
		}

		output, err := kanban.RunScript(kanbanDir, req.Project, slug)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"output": output})
	}
}
