package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"kanban-ui/internal/kanban"
)

func HandleAPIListProjects(kanbanDir string) http.HandlerFunc {
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

func HandleAPIGetProject(kanbanDir string) http.HandlerFunc {
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

func HandleAPIListColumns(kanbanDir string) http.HandlerFunc {
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

func HandleAPIGetTicket(kanbanDir string) http.HandlerFunc {
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

func HandleAPIGetProjectConfig(kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		project := r.PathValue("project")
		cfg, err := kanban.LoadProjectConfig(kanbanDir, project)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, cfg)
	}
}

func HandleAPIUpdateProjectConfig(kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		project := r.PathValue("project")
		var cfg kanban.ProjectConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		// Validate: entries must be non-empty strings
		for _, col := range cfg.ColumnsOrder {
			if strings.TrimSpace(col) == "" {
				writeError(w, http.StatusBadRequest, "column order entries must be non-empty strings")
				return
			}
		}
		if err := kanban.SaveProjectConfig(kanbanDir, project, cfg); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
