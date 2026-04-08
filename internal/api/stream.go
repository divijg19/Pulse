package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/divijg19/Pulse/internal/model"
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

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	clientChan := make(chan model.Event, 10)
	h.Hub.Add(clientChan)
	defer h.Hub.Remove(clientChan)

	for {
		select {
		case event, ok := <-clientChan:
			if !ok {
				return
			}
			jsonData, err := json.Marshal(event.Data)
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(w, "event: %s\n", event.Type); err != nil {
				return
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", jsonData); err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			// The client closed the connection (browser tab closed)
			return
		}
	}
}
