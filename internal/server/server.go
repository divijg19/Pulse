package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/divijg19/Pulse/internal/api"
	"github.com/divijg19/Pulse/internal/stream"
)

const DefaultAddr = ":8080"

type Options struct {
	Addr     string
	StaticFS fs.FS
}

func NewHandler(staticFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	if staticFS != nil {
		mux.Handle("/", http.FileServer(http.FS(staticFS)))
	}

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	hub := stream.NewHub()
	runHandler := &api.RunHandler{Hub: hub}
	mux.HandleFunc("/run", runHandler.HandleRun)
	mux.Handle("/stream", &api.StreamHandler{Hub: hub})

	return mux
}

func New(opts Options) *http.Server {
	addr := opts.Addr
	if addr == "" {
		addr = DefaultAddr
	}

	return &http.Server{
		Addr:              addr,
		Handler:           NewHandler(opts.StaticFS),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func ListenAndServe(opts Options) error {
	return New(opts).ListenAndServe()
}
