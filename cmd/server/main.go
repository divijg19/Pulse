package main

import (
	"fmt"
	"net/http"

	"github.com/divijg19/Pulse/internal/api"
	"github.com/divijg19/Pulse/internal/stream"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})
	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)

	hub := stream.NewHub()

	http.Handle("/stream", &api.StreamHandler{Hub: hub})
}
