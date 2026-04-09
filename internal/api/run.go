package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/divijg19/Pulse/internal/engine"
	"github.com/divijg19/Pulse/internal/model"
	"github.com/divijg19/Pulse/internal/stream"
)

type RunHandler struct {
	Hub *stream.Hub
}

const maxRunRequestBodyBytes = 1 << 20

func (h *RunHandler) HandleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, maxRunRequestBodyBytes)

	req := model.RunRequest{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Concurrency <= 0 || req.Concurrency > 100 {
		http.Error(w, "Concurrency must be between 1 and 100", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(req); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Received run request: %+v\n", req)

	engine.ExecuteConcurrent(r.Context(), req, h.Hub)
	fmt.Printf("Finished executing %d requests to %s with method %s\n", req.Concurrency, req.URL, req.Method)
}
