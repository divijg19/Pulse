package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/divijg19/Pulse/internal/api"
	"github.com/divijg19/Pulse/internal/stream"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("Failed to load embedded static files: %v", err)
	}

	mux := http.NewServeMux()

	// Serve the embedded files
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	hub := stream.NewHub()

	runHandler := &api.RunHandler{Hub: hub}
	mux.HandleFunc("/run", runHandler.HandleRun)
	mux.Handle("/stream", &api.StreamHandler{Hub: hub})

	fmt.Println("⚡ Pulse Engine Running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
