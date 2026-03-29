package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/divijg19/Pulse/internal/stream"
)

type StreamHandler struct {
	Hub *stream.Hub
}

func (h *StreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := make(chan any, 10)

	// NOTE: we'll fix typing later — focus on flow first

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	for {
		select {
		case msg := <-ch:
			data, _ := json.Marshal(msg)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
