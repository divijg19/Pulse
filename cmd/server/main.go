package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/divijg19/Pulse/internal/api"
	"github.com/divijg19/Pulse/internal/model"
	"github.com/divijg19/Pulse/internal/stream"
)

func main() {
	fs := http.FileServer(http.Dir("static/"))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	hub := stream.NewHub()

	mux.HandleFunc("/run", api.HandleRun)
	mux.Handle("/stream", &api.StreamHandler{Hub: hub})

	go func() {
		for {
			time.Sleep(2 * time.Second)
			hub.Broadcast(model.Event{Type: "ping", Data: "hello from hub!"})
		}
	}()

	fmt.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
