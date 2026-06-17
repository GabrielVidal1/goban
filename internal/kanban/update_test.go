package kanban

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestUpdateTicketFieldNewField guards against the corruption bug where setting
// a field that wasn't already present injected a bogus "title: placeholder"
// line and duplicated the title.
func TestUpdateTicketFieldNewField(t *testing.T) {
	dir := t.TempDir()
	col := filepath.Join(dir, "proj", "Done")
	if err := os.MkdirAll(col, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(col, "demo-1.md")
	orig := "---\ntitle: Demo\ncreated: 2026-06-17\n---\nbody text\n"
	if err := os.WriteFile(path, []byte(orig), 0644); err != nil {
		t.Fatal(err)
	}

	// Set two previously-absent fields.
	if _, err := UpdateTicketField(dir, "proj", "demo-1", "priority", "high"); err != nil {
		t.Fatal(err)
	}
	if _, err := UpdateTicketField(dir, "proj", "demo-1", "tags", "a,b"); err != nil {
		t.Fatal(err)
	}

	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)

	if strings.Contains(got, "placeholder") {
		t.Errorf("front matter contains fabricated placeholder:\n%s", got)
	}
	if n := strings.Count(got, "title:"); n != 1 {
		t.Errorf("expected exactly 1 title line, got %d:\n%s", n, got)
	}

	ticket, err := ParseTicket(path)
	if err != nil {
		t.Fatal(err)
	}
	if ticket.Title != "Demo" {
		t.Errorf("Title = %q, want %q", ticket.Title, "Demo")
	}
	if ticket.Priority != "high" {
		t.Errorf("Priority = %q, want %q", ticket.Priority, "high")
	}
	if ticket.Body != "body text" {
		t.Errorf("Body = %q, want %q", ticket.Body, "body text")
	}
}

// TestCreateParseRoundtrip guards ParseTicket's body extraction (previously an
// off-by-one that leaked a fence dash into the body).
func TestCreateParseRoundtrip(t *testing.T) {
	dir := t.TempDir()
	tk, err := CreateTicket(dir, "proj", "Done", "My Ticket", map[string]string{
		"priority": "high",
		"tags":     "a,b",
		"body":     "line one\nline two",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseTicket(tk.Path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "My Ticket" {
		t.Errorf("Title = %q, want %q", got.Title, "My Ticket")
	}
	if got.Body != "line one\nline two" {
		t.Errorf("Body = %q, want %q", got.Body, "line one\nline two")
	}
	if got.Priority != "high" {
		t.Errorf("Priority = %q, want %q", got.Priority, "high")
	}
	if len(got.Tags) != 2 || got.Tags[0] != "a" || got.Tags[1] != "b" {
		t.Errorf("Tags = %v, want [a b]", got.Tags)
	}
}

// TestUpdateTicketFieldExisting confirms updating an existing field replaces it
// in place rather than appending a duplicate.
func TestUpdateTicketFieldExisting(t *testing.T) {
	dir := t.TempDir()
	col := filepath.Join(dir, "proj", "Done")
	if err := os.MkdirAll(col, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(col, "demo-1.md")
	orig := "---\ntitle: Demo\npriority: low\ncreated: 2026-06-17\n---\nbody\n"
	if err := os.WriteFile(path, []byte(orig), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := UpdateTicketField(dir, "proj", "demo-1", "priority", "high"); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(path)
	if n := strings.Count(string(out), "priority:"); n != 1 {
		t.Errorf("expected exactly 1 priority line, got %d:\n%s", n, string(out))
	}
	ticket, _ := ParseTicket(path)
	if ticket.Priority != "high" {
		t.Errorf("Priority = %q, want %q", ticket.Priority, "high")
	}
}
