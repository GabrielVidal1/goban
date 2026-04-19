package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"kanban-ui/internal/config"
	"kanban-ui/internal/kanban"
)

var globalConfig *config.Config

// Run is the unified entry point for the CLI. It parses global flags, loads
// config, and dispatches to the appropriate command. Returns a process exit code.
func Run(args []string) int {
	if len(args) == 0 {
		args = []string{"serve"}
	}

	globalFlags, rest := ParseTopLevel(args)

	// Build config from global flags + env + .env defaults.
	globalConfig = config.Load(globalFlags.Dir, globalFlags.Host, globalFlags.Port)

	if len(rest) == 0 {
		rest = []string{"serve"}
	}
	top := rest[0]
	cmdArgs := rest[1:]

	switch top {
	case "serve":
		return cmdServe(cmdArgs)
	case "help", "-h", "--help":
		printUsage(os.Stdout)
		return 0
	case "projects":
		return dispatch(cmdArgs, map[string]cmdFunc{
			"list": cmdProjectsList,
		})
	case "project":
		return dispatch(cmdArgs, map[string]cmdFunc{
			"info": cmdProjectInfo,
		})
	case "columns":
		return dispatch(cmdArgs, map[string]cmdFunc{
			"list": cmdColumnsList,
		})
	case "tickets":
		return dispatch(cmdArgs, map[string]cmdFunc{
			"list": cmdTicketsList,
		})
	case "ticket":
		return dispatch(cmdArgs, map[string]cmdFunc{
			"get":     cmdTicketGet,
			"create":  cmdTicketCreate,
			"move":    cmdTicketMove,
			"set":     cmdTicketSet,
			"archive": cmdTicketArchive,
			"run":     cmdTicketRun,
		})
	case "config":
		return dispatch(cmdArgs, map[string]cmdFunc{
			"get":       cmdConfigGet,
			"set-order": cmdConfigSetOrder,
		})
	default:
		fail("unknown command: %s", top)
		return 2
	}
}

type cmdFunc func(args []string) int

func dispatch(args []string, subs map[string]cmdFunc) int {
	if len(args) == 0 {
		fail("missing subcommand; expected one of: %s", joinKeys(subs))
		return 2
	}
	fn, ok := subs[args[0]]
	if !ok {
		fail("unknown subcommand: %s (expected one of: %s)", args[0], joinKeys(subs))
		return 2
	}
	return fn(args[1:])
}

func joinKeys(m map[string]cmdFunc) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
}

// ──────────────────────────────── commands ────────────────────────────────

func cmdProjectsList(args []string) int {
	fs := flag.NewFlagSet("projects list", flag.ContinueOnError)
	var dir string
	var asJSON bool
	addCommonFlags(fs, &dir, &asJSON)
	if err := parseFlags(fs, args); err != nil {
		return 2
	}

	projects, err := kanban.ListProjects(dir)
	if err != nil {
		fail("%v", err)
		return 1
	}

	type projectSummary struct {
		Name        string   `json:"name"`
		Columns     []string `json:"columns"`
		TicketCount int      `json:"ticket_count"`
	}
	summaries := make([]projectSummary, 0, len(projects))
	for _, p := range projects {
		info, err := kanban.GetProjectInfo(dir, p)
		if err != nil {
			continue
		}
		summaries = append(summaries, projectSummary{Name: info.Name, Columns: info.Columns, TicketCount: info.TicketCount})
	}

	if asJSON {
		return writeJSON(os.Stdout, summaries)
	}
	tw := tabWriter()
	fmt.Fprintln(tw, "NAME\tCOLUMNS\tTICKETS")
	for _, s := range summaries {
		fmt.Fprintf(tw, "%s\t%s\t%d\n", s.Name, strings.Join(s.Columns, ","), s.TicketCount)
	}
	tw.Flush()
	return 0
}

func cmdProjectInfo(args []string) int {
	fs := flag.NewFlagSet("project info", flag.ContinueOnError)
	var dir string
	var asJSON bool
	addCommonFlags(fs, &dir, &asJSON)
	if err := parseFlags(fs, args); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fail("usage: project info <name>")
		return 2
	}
	info, err := kanban.GetProjectInfo(dir, fs.Arg(0))
	if err != nil {
		fail("%v", err)
		return 1
	}
	if asJSON {
		return writeJSON(os.Stdout, info)
	}
	fmt.Printf("Name:    %s\n", info.Name)
	fmt.Printf("Columns: %s\n", strings.Join(info.Columns, ", "))
	fmt.Printf("Tickets: %d\n", info.TicketCount)
	return 0
}

func cmdColumnsList(args []string) int {
	fs := flag.NewFlagSet("columns list", flag.ContinueOnError)
	var dir string
	var asJSON bool
	addCommonFlags(fs, &dir, &asJSON)
	if err := parseFlags(fs, args); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fail("usage: columns list <project>")
		return 2
	}
	cols, err := kanban.ListColumns(dir, fs.Arg(0))
	if err != nil {
		fail("%v", err)
		return 1
	}
	cfg, _ := kanban.LoadProjectConfig(dir, fs.Arg(0))
	cols = kanban.ApplyColumnOrder(cols, cfg.ColumnsOrder)

	if asJSON {
		return writeJSON(os.Stdout, cols)
	}
	for _, c := range cols {
		fmt.Println(c)
	}
	return 0
}

func cmdTicketsList(args []string) int {
	fs := flag.NewFlagSet("tickets list", flag.ContinueOnError)
	var dir, column, assignee, priority, tag string
	var asJSON bool
	addCommonFlags(fs, &dir, &asJSON)
	fs.StringVar(&column, "column", "", "filter by column")
	fs.StringVar(&assignee, "assignee", "", "filter by assignee")
	fs.StringVar(&priority, "priority", "", "filter by priority")
	fs.StringVar(&tag, "tag", "", "filter by tag")
	if err := parseFlags(fs, args); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fail("usage: tickets list <project> [--column] [--assignee] [--priority] [--tag]")
		return 2
	}
	tickets, err := kanban.ListTicketsFiltered(dir, fs.Arg(0), column, assignee, priority, tag)
	if err != nil {
		fail("%v", err)
		return 1
	}
	if asJSON {
		return writeJSON(os.Stdout, tickets)
	}
	tw := tabWriter()
	fmt.Fprintln(tw, "SLUG\tCOLUMN\tTITLE\tPRIORITY\tASSIGNEE\tDUE")
	for _, t := range tickets {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", t.Slug, t.Column, t.Title, t.Priority, t.Assignee, t.Due)
	}
	tw.Flush()
	return 0
}

func cmdTicketGet(args []string) int {
	fs := flag.NewFlagSet("ticket get", flag.ContinueOnError)
	var dir string
	var asJSON bool
	addCommonFlags(fs, &dir, &asJSON)
	if err := parseFlags(fs, args); err != nil {
		return 2
	}
	if fs.NArg() < 2 {
		fail("usage: ticket get <project> <slug>")
		return 2
	}
	ticket, err := kanban.GetTicket(dir, fs.Arg(0), fs.Arg(1))
	if err != nil {
		fail("%v", err)
		return 1
	}
	if asJSON {
		return writeJSON(os.Stdout, ticket)
	}
	fmt.Printf("Title:    %s\n", ticket.Title)
	fmt.Printf("Slug:     %s\n", ticket.Slug)
	fmt.Printf("Project:  %s\n", ticket.Project)
	fmt.Printf("Column:   %s\n", ticket.Column)
	fmt.Printf("Priority: %s\n", ticket.Priority)
	fmt.Printf("Assignee: %s\n", ticket.Assignee)
	fmt.Printf("Due:      %s\n", ticket.Due)
	fmt.Printf("Tags:     %s\n", strings.Join(ticket.Tags, ", "))
	fmt.Printf("Created:  %s\n", ticket.Created)
	fmt.Printf("Path:     %s\n", ticket.Path)
	if ticket.Body != "" {
		fmt.Printf("\n%s\n", ticket.Body)
	}
	return 0
}

func cmdTicketCreate(args []string) int {
	fs := flag.NewFlagSet("ticket create", flag.ContinueOnError)
	var dir, priority, assignee, due, tags, body string
	addCommonFlags(fs, &dir, nil)
	fs.StringVar(&priority, "priority", "", "")
	fs.StringVar(&assignee, "assignee", "", "")
	fs.StringVar(&due, "due", "", "")
	fs.StringVar(&tags, "tags", "", "comma-separated")
	fs.StringVar(&body, "body", "", "ticket body")
	if err := parseFlags(fs, args); err != nil {
		return 2
	}
	if fs.NArg() < 3 {
		fail("usage: ticket create <project> <column> <title> [--priority] [--assignee] [--due] [--tags] [--body]")
		return 2
	}
	opts := map[string]string{
		"priority": priority,
		"assignee": assignee,
		"due":      due,
		"tags":     tags,
		"body":     body,
	}
	ticket, err := kanban.CreateTicket(dir, fs.Arg(0), fs.Arg(1), fs.Arg(2), opts)
	if err != nil {
		fail("%v", err)
		return 1
	}
	fmt.Printf("Created %s at %s\n", ticket.Slug, ticket.Path)
	return 0
}

func cmdTicketMove(args []string) int {
	fs := flag.NewFlagSet("ticket move", flag.ContinueOnError)
	var dir string
	addCommonFlags(fs, &dir, nil)
	if err := parseFlags(fs, args); err != nil {
		return 2
	}
	if fs.NArg() < 3 {
		fail("usage: ticket move <project> <slug> <column>")
		return 2
	}
	ticket, err := kanban.UpdateTicketStatus(dir, fs.Arg(0), fs.Arg(1), fs.Arg(2))
	if err != nil {
		fail("%v", err)
		return 1
	}
	fmt.Printf("Moved %s to %s\n", ticket.Slug, ticket.Column)
	return 0
}

func cmdTicketSet(args []string) int {
	fs := flag.NewFlagSet("ticket set", flag.ContinueOnError)
	var dir string
	addCommonFlags(fs, &dir, nil)
	if err := parseFlags(fs, args); err != nil {
		return 2
	}
	if fs.NArg() < 4 {
		fail("usage: ticket set <project> <slug> <field> <value>")
		return 2
	}
	ticket, err := kanban.UpdateTicketField(dir, fs.Arg(0), fs.Arg(1), fs.Arg(2), fs.Arg(3))
	if err != nil {
		fail("%v", err)
		return 1
	}
	fmt.Printf("Updated %s: %s = %s\n", ticket.Slug, fs.Arg(2), fs.Arg(3))
	return 0
}

func cmdTicketArchive(args []string) int {
	fs := flag.NewFlagSet("ticket archive", flag.ContinueOnError)
	var dir string
	var del bool
	addCommonFlags(fs, &dir, nil)
	fs.BoolVar(&del, "delete", false, "delete instead of archiving")
	if err := parseFlags(fs, args); err != nil {
		return 2
	}
	if fs.NArg() < 2 {
		fail("usage: ticket archive <project> <slug> [--delete]")
		return 2
	}
	if err := kanban.ArchiveTicket(dir, fs.Arg(0), fs.Arg(1), del); err != nil {
		fail("%v", err)
		return 1
	}
	if del {
		fmt.Printf("Deleted %s\n", fs.Arg(1))
	} else {
		fmt.Printf("Archived %s\n", fs.Arg(1))
	}
	return 0
}

func cmdTicketRun(args []string) int {
	fs := flag.NewFlagSet("ticket run", flag.ContinueOnError)
	var dir string
	addCommonFlags(fs, &dir, nil)
	if err := parseFlags(fs, args); err != nil {
		return 2
	}
	if fs.NArg() < 2 {
		fail("usage: ticket run <project> <slug>")
		return 2
	}
	out, err := kanban.RunScript(dir, fs.Arg(0), fs.Arg(1))
	if err != nil {
		fail("%v", err)
		return 1
	}
	fmt.Print(out)
	return 0
}

func cmdConfigGet(args []string) int {
	fs := flag.NewFlagSet("config get", flag.ContinueOnError)
	var dir string
	var asJSON bool
	addCommonFlags(fs, &dir, &asJSON)
	if err := parseFlags(fs, args); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fail("usage: config get <project>")
		return 2
	}
	cfg, err := kanban.LoadProjectConfig(dir, fs.Arg(0))
	if err != nil {
		fail("%v", err)
		return 1
	}
	if asJSON {
		return writeJSON(os.Stdout, cfg)
	}
	fmt.Printf("columnsOrder: %s\n", strings.Join(cfg.ColumnsOrder, ", "))
	return 0
}

func cmdConfigSetOrder(args []string) int {
	fs := flag.NewFlagSet("config set-order", flag.ContinueOnError)
	var dir string
	addCommonFlags(fs, &dir, nil)
	if err := parseFlags(fs, args); err != nil {
		return 2
	}
	if fs.NArg() < 2 {
		fail("usage: config set-order <project> <col1,col2,...>")
		return 2
	}
	parts := strings.Split(fs.Arg(1), ",")
	order := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			order = append(order, s)
		}
	}
	cfg := kanban.ProjectConfig{ColumnsOrder: order}
	if err := kanban.SaveProjectConfig(dir, fs.Arg(0), cfg); err != nil {
		fail("%v", err)
		return 1
	}
	fmt.Printf("Saved column order for %s\n", fs.Arg(0))
	return 0
}

// ──────────────────────────────── helpers ────────────────────────────────

func tabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
}

func writeJSON(w io.Writer, v any) int {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fail("%v", err)
		return 1
	}
	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `Usage: kanban-ui [global flags] <command> [subcommand] [flags] [args]

Server
  serve                                             Start the web UI server (default when no command given)

Projects
  projects list                                     List all projects
  project info <name>                               Show columns and ticket count
  columns list <project>                            List columns (in configured order)

Tickets
  tickets list <project> [--column] [--assignee] [--priority] [--tag]
  ticket get <project> <slug>
  ticket create <project> <column> <title> [--priority] [--assignee] [--due] [--tags] [--body]
  ticket move <project> <slug> <column>
  ticket set <project> <slug> <field> <value>
  ticket archive <project> <slug> [--delete]
  ticket run <project> <slug>                       Execute column's script.sh

Config
  config get <project>
  config set-order <project> <col1,col2,...>

Global flags (apply to all commands)
  --dir <path>    Kanban directory (overrides KANBAN_DIR; default ./kanban)
  --host <host>   Listen host for serve (overrides HOST; default localhost)
  --port <port>   Listen port for serve (overrides PORT; default 8080)

Common flags (data commands)
  --json          Emit JSON (list/get commands only)`)
}
