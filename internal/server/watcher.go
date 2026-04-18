package server

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func StartFileWatcher(kanbanDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create file watcher: %v", err)
		return
	}
	defer watcher.Close()

	go func() {
		for event := range watcher.Events {
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 &&
				strings.HasSuffix(event.Name, ".md") {
				log.Printf("File changed: %s", event.Name)
				sseClientsMu.Lock()
				alive := sseClients[:0]
				for _, ch := range sseClients {
					select {
					case ch <- "refresh":
						alive = append(alive, ch)
					default:
						close(ch)
					}
				}
				sseClients = alive
				sseClientsMu.Unlock()
			}
		}
	}()

	filepath.Walk(kanbanDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			watcher.Add(path)
		}
		return nil
	})

	select {}
}
