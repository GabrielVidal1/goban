package server

import (
	"io"
	"io/fs"
	"net/http"
	"strings"
)

func SPAHandler(distFS fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(distFS))
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		f, err := distFS.Open(path)
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		// Fall through to index.html for React Router client-side routes
		index, err := distFS.Open("index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer index.Close()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.Copy(w, index)
	}
}
