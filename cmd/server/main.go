package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

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

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	fmt.Println("⚡ Pulse Engine Running on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
