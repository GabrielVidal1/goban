package main

import (
	"embed"
	"io/fs"
	"log"
	"os"

	"kanban-ui/internal/cli"
)

//go:embed ui/dist
var uiFiles embed.FS

func main() {
	distFS, err := fs.Sub(uiFiles, "ui/dist")
	if err != nil {
		log.Fatalf("Failed to create dist sub-FS: %v", err)
	}
	cli.UIFiles = distFS
	os.Exit(cli.Run(os.Args[1:]))
}
