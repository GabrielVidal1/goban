package server

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type fileEvent struct {
	Op      string `json:"op"`
	Project string `json:"project"`
	Column  string `json:"column"`
	Slug    string `json:"slug"`
	File    string `json:"file"`
}

func StartFileWatcher(kanbanDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create file watcher: %v", err)
		return
	}
	defer watcher.Close()

	filepath.Walk(kanbanDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			watcher.Add(path)
		}
		return nil
	})

	SetWatcherRunning(true)

	go func() {
		for event := range watcher.Events {
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}

			payload := buildPayload(kanbanDir, event)
			if payload == "" {
				continue
			}

			log.Printf("File changed: %s (%s)", event.Name, event.Op)
			sseClientsMu.Lock()
			alive := sseClients[:0]
			for _, ch := range sseClients {
				select {
				case ch <- payload:
					alive = append(alive, ch)
				default:
					close(ch)
				}
			}
			sseClients = alive
			sseClientsMu.Unlock()
		}
	}()

	select {}
}

func buildPayload(kanbanDir string, event fsnotify.Event) string {
	rel, err := filepath.Rel(kanbanDir, event.Name)
	if err != nil {
		return ""
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")

	for _, p := range parts {
		if strings.HasPrefix(p, "_") {
			return ""
		}
	}

	file := parts[len(parts)-1]
	if !strings.HasSuffix(file, ".md") {
		return ""
	}

	var project, column string
	if len(parts) >= 3 {
		project = parts[0]
		column = parts[len(parts)-2]
	} else {
		return ""
	}

	op := "write"
	switch {
	case event.Op&fsnotify.Create != 0:
		op = "create"
	case event.Op&(fsnotify.Remove|fsnotify.Rename) != 0:
		op = "remove"
	case event.Op&fsnotify.Write != 0:
		op = "write"
	}

	b, err := json.Marshal(fileEvent{
		Op:      op,
		Project: project,
		Column:  column,
		Slug:    strings.TrimSuffix(file, ".md"),
		File:    file,
	})
	if err != nil {
		return ""
	}
	return string(b)
}
