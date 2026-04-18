package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"kanban-ui/internal/kanban"
)

//go:embed templates static
var embeddedFiles embed.FS

func main() {
	kanbanDir := os.Getenv("KANBAN_DIR")
	if kanbanDir == "" {
		kanbanDir = "./kanban"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	tmpl, err := loadTemplates()
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	mux := http.NewServeMux()

	// Page routes
	mux.HandleFunc("/", handleBoard(tmpl, kanbanDir))
	mux.HandleFunc("/project/", handleProjectBoard(tmpl, kanbanDir))

	// HTMX partials - Board views
	mux.HandleFunc("/board/columns/", handleColumnTickets(tmpl, kanbanDir))
	mux.HandleFunc("/board/ticket/", handleTicketDetail(tmpl, kanbanDir))

	// HTMX action forms
	mux.HandleFunc("/ticket/new/form", handleNewTicketForm(tmpl, kanbanDir))
	mux.HandleFunc("/ticket/move/{slug}/form", handleMoveTicketForm(tmpl, kanbanDir))

	// HTMX action endpoints (POST)
	mux.HandleFunc("/ticket/create", handleCreateTicket(kanbanDir))
	mux.HandleFunc("/ticket/move/", handleMoveTicket(kanbanDir))
	mux.HandleFunc("/ticket/update-field/", handleUpdateField(kanbanDir))
	mux.HandleFunc("/ticket/archive/{slug}", handleArchiveTicket(kanbanDir))

	// Static files
	staticFS, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Fatalf("Failed to create static sub-FS: %v", err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	log.Printf("Kanban UI starting on :%s (KANBAN_DIR=%s)", port, kanbanDir)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func loadTemplates() (*template.Template, error) {
	funcMap := template.FuncMap{
		"upper":    strings.ToUpper,
		"lower":    strings.ToLower,
		"title":    strings.ToTitle,
		"firstUpper": func(s string) string {
			if s == "" { return s }
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"firstLower": func(s string) string {
			if s == "" { return s }
			return strings.ToLower(s[:1]) + s[1:]
		},
		"truncate": func(s string, n int) string {
			if len(s) <= n { return s }
			return s[:n] + "..."
		},
		"priorityClass": func(p string) string {
			switch strings.ToLower(p) {
			case "critical", "blocker": return "priority-critical"
			case "high": return "priority-high"
			case "medium": return "priority-medium"
			case "low": return "priority-low"
			default: return ""
			}
		},
		"tagClass": func(t string) string {
			switch strings.ToLower(t) {
			case "bug", "fix": return "tag-bug"
			case "feature", "feat": return "tag-feature"
			case "improvement", "enhancement": return "tag-improvement"
			default: return "tag-default"
			}
		},
		"isOverdue": func(due string) bool {
			if due == "" { return false }
			d, err := time.Parse("2006-01-02", due)
			if err != nil { return false }
			return d.Before(time.Now())
		},
	}

	tmpl := template.New("").Funcs(funcMap)

	templateFiles, err := fs.Sub(embeddedFiles, "templates")
	if err != nil {
		return nil, fmt.Errorf("failed to create templates sub-FS: %w", err)
	}

	err = fs.WalkDir(templateFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".html") { return nil }
		data, err := embeddedFiles.ReadFile("templates/" + path)
		if err != nil { return err }
		_, err = tmpl.New(path).Parse(string(data))
		return err
	})

	return tmpl, err
}

// ---- Page Handlers ----

func handleBoard(tmpl *template.Template, kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projects, _ := kanban.ListProjects(kanbanDir)

		type ProjectSummary struct {
			Name        string
			ColumnCount int
			TicketCount int
		}

		var summaries []ProjectSummary
		for _, p := range projects {
			info, err := kanban.GetProjectInfo(kanbanDir, p)
			if err != nil { continue }
			summaries = append(summaries, ProjectSummary{
				Name: info.Name, ColumnCount: len(info.Columns), TicketCount: info.TicketCount,
			})
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "board.html", map[string]interface{}{
			"Projects": summaries, "KanbanDir": kanbanDir,
		})
	}
}

func handleProjectBoard(tmpl *template.Template, kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectName := strings.TrimPrefix(r.URL.Path, "/project/")
		if projectName == "" || projectName == "/" { return }

		board, err := kanban.GetBoardData(kanbanDir, projectName)
		if err != nil {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "board.html", map[string]interface{}{
			"CurrentProject": projectName, "Board": board,
			"AllProjects": getProjectList(kanbanDir), "KanbanDir": kanbanDir,
		})
	}
}

func handleColumnTickets(tmpl *template.Template, kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/board/columns/")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) < 2 || parts[1] != "tickets" { return }

		projectName, columnName := parts[0], parts[1]
		board, err := kanban.GetBoardData(kanbanDir, projectName)
		if err != nil { http.Error(w, "Project not found", http.StatusNotFound); return }

		var tickets []kanban.Ticket
		for _, col := range board.Columns {
			if col.Name == columnName { tickets = col.Tickets; break }
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "partials/ticket-list.html", map[string]interface{}{
			"Project": projectName, "Column": columnName, "Tickets": tickets,
			"AllProjects": getProjectList(kanbanDir),
		})
	}
}

func handleTicketDetail(tmpl *template.Template, kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/board/ticket/")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) < 2 { return }

		projectName, slug := parts[0], parts[1]
		ticket, err := kanban.GetTicket(kanbanDir, projectName, slug)
		if err != nil { http.Error(w, "Ticket not found", http.StatusNotFound); return }

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "partials/ticket-detail.html", map[string]interface{}{
			"Project": projectName, "Ticket": ticket,
			"AllProjects": getProjectList(kanbanDir),
		})
	}
}

// ---- Action Handlers (HTMX) ----

func handleNewTicketForm(tmpl *template.Template, kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectName := r.URL.Query().Get("project")

		var columns []string
		if projectName != "" {
			cols, _ := kanban.ListColumns(kanbanDir, projectName)
			columns = cols
		} else {
			type ColInfo struct{ Project, Name string }
			var allCols []ColInfo
			projects, _ := kanban.ListProjects(kanbanDir)
			for _, p := range projects {
				cols, _ := kanban.ListColumns(kanbanDir, p)
				for _, c := range cols { allCols = append(allCols, ColInfo{Project: p, Name: c}) }
			}
			columns = make([]string, 0, len(allCols))
			for _, c := range allCols { columns = append(columns, fmt.Sprintf("%s/%s", c.Project, c.Name)) }
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "partials/new-ticket-form.html", map[string]interface{}{
			"Project": projectName, "Columns": columns,
			"AllProjects": getProjectList(kanbanDir),
		})
	}
}

func handleCreateTicket(kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		column := r.FormValue("column")
		title := r.FormValue("title")
		priority := r.FormValue("priority")
		assignee := r.FormValue("assignee")
		due := r.FormValue("due")
		tags := r.FormValue("tags")
		body := r.FormValue("body")

		if title == "" { http.Error(w, "Title is required", http.StatusBadRequest); return }

		var project string
		parts := strings.SplitN(column, "/", 2)
		if len(parts) == 2 {
			project = parts[0]
			column = parts[1]
		} else if r.FormValue("project") != "" {
			project = r.FormValue("project")
		}

		opts := map[string]string{}
		if priority != "" { opts["priority"] = priority }
		if assignee != "" { opts["assignee"] = assignee }
		if due != "" { opts["due"] = due }
		if tags != "" { opts["tags"] = tags }
		if body != "" { opts["body"] = body }

		ticket, err := kanban.CreateTicket(kanbanDir, project, column, title, opts)
		if err != nil { http.Error(w, "Failed to create ticket: "+err.Error(), http.StatusInternalServerError); return }

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<div class="ticket-card" hx-get="/board/ticket/%s/%s" hx-target="#modal-content" hx-swap="innerHTML">
			<div class="ticket-header"><span class="ticket-title">%s</span>%s</div>
			<div class="ticket-meta">%v</div>
		</div>`, project, ticket.Slug, escapeHTML(ticket.Title), priorityBadge(ticket.Priority), tagBadges(ticket.Tags))
	}
}

func handleMoveTicketForm(tmpl *template.Template, kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		projectName := r.URL.Query().Get("project")

		ticket, err := kanban.GetTicket(kanbanDir, projectName, slug)
		if err != nil { http.Error(w, "Ticket not found", http.StatusNotFound); return }

		columns, _ := kanban.ListColumns(kanbanDir, projectName)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "partials/move-ticket-form.html", map[string]interface{}{
			"Project": projectName, "Ticket": ticket, "Columns": columns,
			"AllProjects": getProjectList(kanbanDir),
		})
	}
}

func handleMoveTicket(kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		slug := r.FormValue("slug")
		targetColumn := r.FormValue("target_column")
		projectName := r.FormValue("project")

		if targetColumn == "" || projectName == "" { http.Error(w, "Missing required fields", http.StatusBadRequest); return }

		ticket, err := kanban.UpdateTicketStatus(kanbanDir, projectName, slug, targetColumn)
		if err != nil { http.Error(w, "Failed to move ticket: "+err.Error(), http.StatusInternalServerError); return }

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<div class="alert alert-success">Moved "%s" to %s</div>`, escapeHTML(ticket.Title), escapeHTML(targetColumn))
	}
}

func handleUpdateField(kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		slug := r.FormValue("slug")
		field := r.FormValue("field")
		value := r.FormValue("value")
		projectName := r.FormValue("project")

		if field == "" || projectName == "" { http.Error(w, "Missing required fields", http.StatusBadRequest); return }

		ticket, err := kanban.UpdateTicketField(kanbanDir, projectName, slug, field, value)
		if err != nil { http.Error(w, "Failed to update: "+err.Error(), http.StatusInternalServerError); return }

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<span class="field-value" data-field="%s">%s</span>`, field, formatFieldValue(field, ticket))
	}
}

func handleArchiveTicket(kanbanDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		projectName := r.URL.Query().Get("project")

		if projectName == "" { http.Error(w, "Missing project", http.StatusBadRequest); return }

		err := kanban.ArchiveTicket(kanbanDir, projectName, slug, false)
		if err != nil { http.Error(w, "Failed to archive: "+err.Error(), http.StatusInternalServerError); return }

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<div class="alert alert-success">Ticket archived successfully</div>`)
	}
}

// ---- Helpers ----

func getProjectList(kanbanDir string) []string {
	projects, err := kanban.ListProjects(kanbanDir)
	if err != nil { return []string{} }
	return projects
}

func escapeHTML(s string) template.HTML {
	return template.HTML(template.HTMLEscapeString(s))
}

func priorityBadge(priority string) template.HTML {
	if priority == "" { return "" }
	class := "priority-" + strings.ToLower(priority)
	name := strings.ToUpper(priority[:1]) + strings.ToLower(priority[1:])
	return template.HTML(`<span class="badge badge-sm ` + class + `">` + name + `</span>`)
}

func tagBadges(tags []string) template.HTML {
	if len(tags) == 0 { return "" }
	var parts []string
	for _, t := range tags {
		parts = append(parts, `<span class="badge badge-sm tag-default">`+template.HTMLEscapeString(t)+`</span>`)
	}
	return template.HTML(strings.Join(parts, " "))
}

func formatFieldValue(field string, ticket *kanban.Ticket) string {
	switch strings.ToLower(field) {
	case "priority":
		if ticket.Priority == "" { return "<em>none</em>" }
		return string(escapeHTML(ticket.Priority))
	case "assignee":
		if ticket.Assignee == "" { return "<em>unassigned</em>" }
		return string(escapeHTML(ticket.Assignee))
	case "due":
		if ticket.Due == "" { return "<em>none</em>" }
		return string(escapeHTML(ticket.Due))
	case "tags":
		if len(ticket.Tags) == 0 { return "<em>none</em>" }
		return string(escapeHTML(strings.Join(ticket.Tags, ", ")))
	default: return ""
	}
}
