package server

import (
	"net/http"
	"sync"
)

var (
	sseClients   []chan string
	sseClientsMu sync.Mutex
)

func HandleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan string, 10)
	sseClientsMu.Lock()
	sseClients = append(sseClients, ch)
	sseClientsMu.Unlock()

	defer func() {
		sseClientsMu.Lock()
		for i, c := range sseClients {
			if c == ch {
				sseClients = append(sseClients[:i], sseClients[i+1:]...)
				close(ch)
				break
			}
		}
		sseClientsMu.Unlock()
	}()

	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-ch:
			w.Write([]byte("event: filechange\ndata: " + msg + "\n\n"))
			flusher.Flush()
		}
	}
}
