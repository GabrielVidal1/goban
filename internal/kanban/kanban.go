package kanban

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Ticket struct {
	Title    string   `json:"title"`
	Priority string   `json:"priority"`
	Assignee string   `json:"assignee"`
	Due      string   `json:"due"`
	Tags     []string `json:"tags"`
	Created  string   `json:"created"`
	Body     string   `json:"body"`
	Slug     string   `json:"slug"`
	Column   string   `json:"column"`
	Project  string   `json:"project"`
	Path     string   `json:"path"`
}

type Project struct {
	Name        string   `json:"name"`
	Columns     []string `json:"columns"`
	TicketCount int      `json:"ticket_count"`
}

const defaultKanbanDir = "./kanban"

func ResolveProjectDir(kanbanDir, projectName string) (string, error) {
	if kanbanDir == "" {
		kanbanDir = defaultKanbanDir
	}
	if _, err := os.Stat(kanbanDir); os.IsNotExist(err) {
		return "", fmt.Errorf("kanban directory not found at '%s'", kanbanDir)
	}

	entries, _ := os.ReadDir(kanbanDir)
	var projects []string
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), "_") {
			continue
		}
		subEntries, _ := os.ReadDir(filepath.Join(kanbanDir, entry.Name()))
		hasSubDirs := false
		for _, se := range subEntries {
			if se.IsDir() {
				hasSubDirs = true
				break
			}
		}
		if hasSubDirs {
			projects = append(projects, entry.Name())
		}
	}
	sort.Strings(projects)

	if projectName != "" {
		target := filepath.Join(kanbanDir, projectName)
		if _, err := os.Stat(target); os.IsNotExist(err) {
			return "", fmt.Errorf("project '%s' not found in %s", projectName, kanbanDir)
		}
		return target, nil
	}
	if len(projects) == 0 {
		return "", fmt.Errorf("no projects found in %s", kanbanDir)
	}
	if len(projects) == 1 {
		return filepath.Join(kanbanDir, projects[0]), nil
	}
	return "", fmt.Errorf("multiple projects — specify one: %s", strings.Join(projects, ", "))
}

func ListProjects(kanbanDir string) ([]string, error) {
	if kanbanDir == "" {
		kanbanDir = defaultKanbanDir
	}
	entries, _ := os.ReadDir(kanbanDir)
	var projects []string
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), "_") {
			continue
		}
		subEntries, _ := os.ReadDir(filepath.Join(kanbanDir, entry.Name()))
		hasSubDirs := false
		for _, se := range subEntries {
			if se.IsDir() {
				hasSubDirs = true
				break
			}
		}
		if hasSubDirs {
			projects = append(projects, entry.Name())
		}
	}
	sort.Strings(projects)
	return projects, nil
}

func ListColumns(kanbanDir, projectName string) ([]string, error) {
	if kanbanDir == "" {
		kanbanDir = defaultKanbanDir
	}
	projDir := filepath.Join(kanbanDir, projectName)
	entries, err := os.ReadDir(projDir)
	if err != nil {
		return nil, err
	}
	var columns []string
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), "_") {
			continue
		}
		columns = append(columns, entry.Name())
	}
	sort.Strings(columns)
	return columns, nil
}

func ListTickets(kanbanDir, projectName string) ([]Ticket, error) {
	if kanbanDir == "" {
		kanbanDir = defaultKanbanDir
	}
	projDir := filepath.Join(kanbanDir, projectName)
	var tickets []Ticket
	err := filepath.Walk(projDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		ticket, e := ParseTicket(path)
		if e == nil {
			tickets = append(tickets, ticket)
		}
		return nil
	})
	sort.Slice(tickets, func(i, j int) bool { return tickets[i].Title < tickets[j].Title })
	return tickets, err
}

func ListTicketsFiltered(kanbanDir, projectName, column, assignee, priority, tag string) ([]Ticket, error) {
	tickets, err := ListTickets(kanbanDir, projectName)
	if err != nil {
		return nil, err
	}
	var filtered []Ticket
	for _, t := range tickets {
		if column != "" && t.Column != column {
			continue
		}
		if assignee != "" && !strings.EqualFold(t.Assignee, assignee) {
			continue
		}
		if priority != "" && !strings.EqualFold(t.Priority, priority) {
			continue
		}
		if tag != "" {
			found := false
			for _, tg := range t.Tags {
				if strings.EqualFold(tg, tag) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		filtered = append(filtered, t)
	}
	return filtered, nil
}

func GetTicket(kanbanDir, projectName, slug string) (*Ticket, error) {
	if kanbanDir == "" {
		kanbanDir = defaultKanbanDir
	}
	projDir := filepath.Join(kanbanDir, projectName)
	var matched *Ticket
	err := filepath.Walk(projDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		nameLower := strings.ToLower(info.Name())
		slugLower := strings.ToLower(slug)
		if strings.Contains(nameLower, slugLower) {
			ticket, e := ParseTicket(path)
			if e == nil {
				matched = &ticket
				return filepath.SkipAll
			}
		}
		return nil
	})
	if matched == nil {
		return nil, fmt.Errorf("ticket '%s' not found in %s", slug, projDir)
	}
	return matched, err
}

func ParseTicket(path string) (Ticket, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Ticket{}, err
	}
	content := strings.TrimSpace(string(data))
	var ticket Ticket
	ticket.Path = path
	filename := filepath.Base(path)
	ticket.Slug = strings.TrimSuffix(filename, ".md")

	if strings.HasPrefix(content, "---") {
		endIdx := strings.Index(content[3:], "\n---")
		fmContent := ""
		body := content
		if endIdx > 0 {
			fmContent = content[4 : 3+endIdx]
			body = strings.TrimSpace(content[6+endIdx:])
		}
		ticket.Body = body
		for _, line := range strings.Split(fmContent, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			colonIdx := strings.Index(line, ":")
			if colonIdx < 0 {
				continue
			}
			key := strings.ToLower(strings.TrimSpace(line[:colonIdx]))
			val := strings.TrimSpace(strings.Trim(line[colonIdx+1:], `"'`))
			switch key {
			case "title":
				ticket.Title = val
			case "priority":
				ticket.Priority = val
			case "assignee":
				ticket.Assignee = val
			case "due":
				ticket.Due = val
			case "tags":
				val = strings.Trim(val, "[]")
				for _, tag := range strings.Split(val, ",") {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						ticket.Tags = append(ticket.Tags, tag)
					}
				}
			case "created":
				ticket.Created = val
			}
		}
	} else {
		ticket.Body = content
	}

	dir := filepath.Dir(path)
	ticket.Column = filepath.Base(dir)
	ticket.Project = filepath.Base(filepath.Dir(dir))
	return ticket, nil
}

func CreateTicket(kanbanDir, projectName, column, title string, opts map[string]string) (*Ticket, error) {
	if kanbanDir == "" {
		kanbanDir = defaultKanbanDir
	}
	projDir := filepath.Join(kanbanDir, projectName)
	colDir := filepath.Join(projDir, column)
	if err := os.MkdirAll(colDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", colDir, err)
	}

	slug := strings.ToLower(title)
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%s-%d.md", slug, timestamp)
	path := filepath.Join(colDir, filename)

	var fmLines []string
	fmLines = append(fmLines, "---")
	fmLines = append(fmLines, "title: "+title)
	if v, ok := opts["priority"]; ok && v != "" {
		fmLines = append(fmLines, "priority: "+v)
	}
	if v, ok := opts["assignee"]; ok && v != "" {
		fmLines = append(fmLines, "assignee: "+v)
	}
	if v, ok := opts["due"]; ok && v != "" {
		fmLines = append(fmLines, "due: "+v)
	}
	if v, ok := opts["tags"]; ok && v != "" {
		fmLines = append(fmLines, "tags: ["+strings.ReplaceAll(v, ", ", ",")+"]")
	}
	fmLines = append(fmLines, "created: "+time.Now().Format("2006-01-02"))
	fmLines = append(fmLines, "---")

	body := ""
	if b, ok := opts["body"]; ok {
		body = b
	}
	content := strings.Join(fmLines, "\n") + "\n" + body + "\n"

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write ticket: %w", err)
	}
	ticket, _ := ParseTicket(path)
	return &ticket, nil
}

func UpdateTicketStatus(kanbanDir, projectName, slug, targetColumn string) (*Ticket, error) {
	if kanbanDir == "" {
		kanbanDir = defaultKanbanDir
	}
	projDir := filepath.Join(kanbanDir, projectName)
	var oldPath string
	err := filepath.Walk(projDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		nameLower := strings.ToLower(info.Name())
		slugLower := strings.ToLower(slug)
		if strings.Contains(nameLower, slugLower) {
			oldPath = path
			return filepath.SkipAll
		}
		return nil
	})
	if oldPath == "" {
		return nil, fmt.Errorf("ticket '%s' not found in %s", slug, projDir)
	}

	ticket, err := ParseTicket(oldPath)
	if err != nil {
		return nil, err
	}

	targetColDir := filepath.Join(projDir, targetColumn)
	if err := os.MkdirAll(targetColDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", targetColDir, err)
	}
	newPath := filepath.Join(targetColDir, filepath.Base(oldPath))
	if err := os.Rename(oldPath, newPath); err != nil {
		return nil, fmt.Errorf("failed to move ticket: %w", err)
	}

	ticket.Column = targetColumn
	ticket.Path = newPath
	return &ticket, nil
}

func UpdateTicketField(kanbanDir, projectName, slug, field, value string) (*Ticket, error) {
	if kanbanDir == "" {
		kanbanDir = defaultKanbanDir
	}
	projDir := filepath.Join(kanbanDir, projectName)
	var ticketPath string
	_ = filepath.Walk(projDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		nameLower := strings.ToLower(info.Name())
		slugLower := strings.ToLower(slug)
		if strings.Contains(nameLower, slugLower) {
			ticketPath = path
			return filepath.SkipAll
		}
		return nil
	})
	if ticketPath == "" {
		return nil, fmt.Errorf("ticket '%s' not found in %s", slug, projDir)
	}

	data, err := os.ReadFile(ticketPath)
	if err != nil {
		return nil, err
	}
	content := strings.TrimSpace(string(data))
	fieldLower := strings.ToLower(field)

	var newContent string
	if strings.HasPrefix(content, "---") {
		endIdx := strings.Index(content[3:], "\n---")
		if endIdx > 0 {
			fmContent := content[4 : 3+endIdx]
			body := content[6+endIdx:]
			var newFM []string
			fieldUpdated := false
			for _, line := range strings.Split(fmContent, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					newFM = append(newFM, line)
					continue
				}
				colonIdx := strings.Index(line, ":")
				if colonIdx < 0 {
					newFM = append(newFM, line)
					continue
				}
				key := strings.ToLower(strings.TrimSpace(line[:colonIdx]))
				if key == fieldLower {
					newFM = append(newFM, field+": "+value)
					fieldUpdated = true
				} else {
					newFM = append(newFM, line)
				}
			}
			if !fieldUpdated {
				var inserted []string
				inserted = append(inserted, "title: placeholder")
				for _, line := range newFM {
					inserted = append(inserted, line)
					if strings.HasPrefix(line, "title:") && !fieldUpdated {
						inserted = append(inserted, field+": "+value)
						fieldUpdated = true
					}
				}
				newFM = inserted
			}
			fmStr := strings.Join(newFM, "\n")
			newContent = "---\n" + fmStr + "\n---\n" + body
		} else {
			title := extractTitle(content)
			newContent = fmt.Sprintf("---\ntitle: %s\n%s: %s\ncreated: %s\n---\n%s", title, field, value, time.Now().Format("2006-01-02"), content)
		}
	} else {
		title := extractTitle(content)
		newContent = fmt.Sprintf("---\ntitle: %s\n%s: %s\ncreated: %s\n---\n%s", title, field, value, time.Now().Format("2006-01-02"), content)
	}

	if err := os.WriteFile(ticketPath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to update ticket: %w", err)
	}
	ticket, _ := ParseTicket(ticketPath)
	return &ticket, nil
}

func ArchiveTicket(kanbanDir, projectName, slug string, delete bool) error {
	if kanbanDir == "" {
		kanbanDir = defaultKanbanDir
	}
	projDir := filepath.Join(kanbanDir, projectName)
	var ticketPath string
	_ = filepath.Walk(projDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		nameLower := strings.ToLower(info.Name())
		slugLower := strings.ToLower(slug)
		if strings.Contains(nameLower, slugLower) {
			ticketPath = path
			return filepath.SkipAll
		}
		return nil
	})
	if ticketPath == "" {
		return fmt.Errorf("ticket '%s' not found in %s", slug, projDir)
	}

	if delete {
		return os.Remove(ticketPath)
	}
	archiveDir := filepath.Join(projDir, "_archive")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}
	return os.Rename(ticketPath, filepath.Join(archiveDir, filepath.Base(ticketPath)))
}

func GetProjectInfo(kanbanDir, projectName string) (*Project, error) {
	if kanbanDir == "" {
		kanbanDir = defaultKanbanDir
	}
	columns, err := ListColumns(kanbanDir, projectName)
	if err != nil {
		return nil, err
	}
	tickets, _ := ListTickets(kanbanDir, projectName)
	totalTickets := len(tickets)
	return &Project{Name: projectName, Columns: columns, TicketCount: totalTickets}, nil
}

type BoardData struct {
	Project string              `json:"project"`
	Columns []ColumnWithTickets `json:"columns"`
}

type ColumnWithTickets struct {
	Name    string   `json:"name"`
	Tickets []Ticket `json:"tickets"`
}

func GetBoardData(kanbanDir, projectName string) (*BoardData, error) {
	if kanbanDir == "" {
		kanbanDir = defaultKanbanDir
	}
	columns, err := ListColumns(kanbanDir, projectName)
	if err != nil {
		return nil, err
	}
	board := &BoardData{Project: projectName, Columns: make([]ColumnWithTickets, 0, len(columns))}
	for _, col := range columns {
		tickets, _ := ListTicketsFiltered(kanbanDir, projectName, col, "", "", "")
		board.Columns = append(board.Columns, ColumnWithTickets{Name: col, Tickets: tickets})
	}
	return board, nil
}

func extractTitle(content string) string {
	lines := strings.SplitN(content, "\n", 2)
	if len(lines) == 0 {
		return "Untitled"
	}
	title := strings.TrimSpace(lines[0])
	if idx := strings.Index(title, ":"); idx > 0 {
		key := strings.ToLower(strings.TrimSpace(title[:idx]))
		if key != "---" && !strings.Contains(key, " ") {
			title = strings.TrimSpace(title[idx+1:])
		}
	}
	return title
}

// RunScript executes script.sh from the ticket's column folder with the ticket slug as an argument.
func RunScript(kanbanDir, projectName, slug string) (string, error) {
	if kanbanDir == "" {
		kanbanDir = defaultKanbanDir
	}
	projDir := filepath.Join(kanbanDir, projectName)

	var ticketPath string
	err := filepath.Walk(projDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		nameLower := strings.ToLower(info.Name())
		slugLower := strings.ToLower(slug)
		if strings.Contains(nameLower, slugLower) {
			ticketPath = path
			return filepath.SkipAll
		}
		return nil
	})
	if ticketPath == "" {
		return "", fmt.Errorf("ticket '%s' not found in %s", slug, projDir)
	}

	ticket, err := ParseTicket(ticketPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse ticket: %w", err)
	}

	scriptPath := filepath.Join(filepath.Dir(ticketPath), "script.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return "", fmt.Errorf("no script.sh found in column folder '%s'", ticket.Column)
	}

	cmd := exec.Command("bash", scriptPath, slug)
	cmd.Dir = filepath.Join(kanbanDir, projectName, ticket.Column)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("script execution failed: %w\noutput: %s", err, string(output))
	}

	return string(output), nil
}
