package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/divijg19/Pulse/internal/api"
	"github.com/divijg19/Pulse/internal/stream"
)

func main() {
	fs := http.FileServer(http.Dir("static/"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	hub := stream.NewHub()

	mux := http.NewServeMux()
	mux.HandleFunc("/run", api.HandleRun)
	mux.Handle("/stream", &api.StreamHandler{Hub: hub})

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", mux)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
