package kanban

import (
	"encoding/json"
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

type ProjectConfig struct {
	ColumnsOrder []string `json:"columnsOrder"`
	Shortname    string   `json:"shortname,omitempty"`
}

const configFileName = "config.json"

// LoadProjectConfig reads the project's config.json. Returns zero-value struct if file is missing or invalid.
func LoadProjectConfig(kanbanDir, project string) (ProjectConfig, error) {
	if kanbanDir == "" {
		kanbanDir = defaultKanbanDir
	}
	cfgPath := filepath.Join(kanbanDir, project, configFileName)
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ProjectConfig{}, nil
		}
		return ProjectConfig{}, err
	}
	var cfg ProjectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		// Invalid JSON — treat as empty config
		return ProjectConfig{}, nil
	}
	return cfg, nil
}

// SaveProjectConfig writes the project's config.json atomically. Creates parent dir if needed.
func SaveProjectConfig(kanbanDir, project string, cfg ProjectConfig) error {
	if kanbanDir == "" {
		kanbanDir = defaultKanbanDir
	}
	projDir := filepath.Join(kanbanDir, project)
	cfgPath := filepath.Join(projDir, configFileName)
	if err := os.MkdirAll(projDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	tmpPath := cfgPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp config: %w", err)
	}
	if err := os.Rename(tmpPath, cfgPath); err != nil {
		os.Remove(tmpPath) // clean up
		return fmt.Errorf("failed to rename config: %w", err)
	}
	return nil
}

// ApplyColumnOrder returns columns sorted by the given order first (in that order),
// with any unlisted columns appended alphabetically. Entries in order not matching
// actual directories are silently skipped.
func ApplyColumnOrder(columns []string, order []string) []string {
	if len(order) == 0 {
		sort.Strings(columns)
		return columns
	}

	actual := make(map[string]bool)
	for _, c := range columns {
		actual[c] = true
	}

	var ordered []string
	seen := make(map[string]bool)
	for _, o := range order {
		if actual[o] && !seen[o] {
			ordered = append(ordered, o)
			seen[o] = true
		}
	}

	for _, c := range columns {
		if !seen[c] {
			ordered = append(ordered, c)
		}
	}

	return ordered
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

// splitFrontMatter separates a ticket file's YAML-ish front matter from its
// body. It returns the lines between the opening and closing "---" fences
// (exclusive), the body text (with the blank separator after the closing fence
// stripped), and whether a complete front-matter block was found. Line-based so
// it can't slice through a fence the way hand-tuned byte offsets did.
func splitFrontMatter(content string) (fmLines []string, body string, ok bool) {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		return nil, content, false
	}
	lines := strings.Split(content, "\n")
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return lines[1:i], strings.TrimLeft(strings.Join(lines[i+1:], "\n"), "\n"), true
		}
	}
	return nil, content, false
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

	if fmLines, body, ok := splitFrontMatter(content); ok {
		ticket.Body = body
		for _, line := range fmLines {
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

// nextTicketIndex scans all column directories under projDir and returns the
// next available index for the given shortname (max existing index + 1, or 1).
func nextTicketIndex(projDir, shortname string) int {
	prefix := strings.ToUpper(shortname) + "-"
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(prefix) + `(\d+)-`)
	max := 0
	_ = filepath.Walk(projDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		if m := re.FindStringSubmatch(info.Name()); m != nil {
			var n int
			fmt.Sscanf(m[1], "%d", &n)
			if n > max {
				max = n
			}
		}
		return nil
	})
	return max + 1
}

// ValidateShortname returns an error if s is not a valid project shortname.
func ValidateShortname(s string) error {
	if matched, _ := regexp.MatchString(`^[A-Z0-9]{2,6}$`, s); !matched {
		return fmt.Errorf("shortname must be 2–6 uppercase alphanumeric characters (got %q)", s)
	}
	return nil
}

// SetProjectShortname validates and persists a shortname in the project config.
func SetProjectShortname(kanbanDir, project, shortname string) error {
	if err := ValidateShortname(shortname); err != nil {
		return err
	}
	cfg, err := LoadProjectConfig(kanbanDir, project)
	if err != nil {
		return err
	}
	cfg.Shortname = shortname
	return SaveProjectConfig(kanbanDir, project, cfg)
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

	cfg, _ := LoadProjectConfig(kanbanDir, projectName)
	timestamp := time.Now().Unix()
	var filename string
	if cfg.Shortname != "" {
		idx := nextTicketIndex(projDir, cfg.Shortname)
		filename = fmt.Sprintf("%s-%d-%d.md", strings.ToUpper(cfg.Shortname), idx, timestamp)
	} else {
		slug := strings.ToLower(title)
		slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
		slug = strings.Trim(slug, "-")
		filename = fmt.Sprintf("%s-%d.md", slug, timestamp)
	}
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

	fmLines, body, ok := splitFrontMatter(content)
	if !ok {
		// No front matter — synthesize one and keep existing content as body.
		fmLines = []string{"title: " + extractTitle(content), "created: " + time.Now().Format("2006-01-02")}
		body = content
	}

	fieldUpdated := false
	for i, line := range fmLines {
		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			continue
		}
		if strings.ToLower(strings.TrimSpace(line[:colonIdx])) == fieldLower {
			fmLines[i] = field + ": " + value
			fieldUpdated = true
		}
	}
	if !fieldUpdated {
		// Field not present yet — append it rather than fabricating a title.
		fmLines = append(fmLines, field+": "+value)
	}

	newContent := "---\n" + strings.Join(fmLines, "\n") + "\n---\n"
	if body != "" {
		newContent += body + "\n"
	}

	if err := os.WriteFile(ticketPath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to update ticket: %w", err)
	}
	ticket, _ := ParseTicket(ticketPath)
	return &ticket, nil
}

func UpdateTicketBody(kanbanDir, projectName, slug, body string) (*Ticket, error) {
	if kanbanDir == "" {
		kanbanDir = defaultKanbanDir
	}
	projDir := filepath.Join(kanbanDir, projectName)
	var ticketPath string
	_ = filepath.Walk(projDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(slug)) {
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
	var newContent string
	if fmLines, _, ok := splitFrontMatter(content); ok {
		newContent = "---\n" + strings.Join(fmLines, "\n") + "\n---\n" + strings.TrimLeft(body, "\n") + "\n"
	} else {
		newContent = fmt.Sprintf("---\ntitle: %s\ncreated: %s\n---\n%s\n", extractTitle(content), time.Now().Format("2006-01-02"), body)
	}
	if err := os.WriteFile(ticketPath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to update ticket body: %w", err)
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
	cfg, _ := LoadProjectConfig(kanbanDir, projectName)
	columns = ApplyColumnOrder(columns, cfg.ColumnsOrder)
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
	cfg, _ := LoadProjectConfig(kanbanDir, projectName)
	columns = ApplyColumnOrder(columns, cfg.ColumnsOrder)
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
