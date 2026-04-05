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
		case event := <-clientChan:
			jsonData, _ := json.Marshal(event.Data)
			fmt.Fprintf(w, "event: %s\n", event.Type)
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			flusher.Flush()
		case <-r.Context().Done():
			// The client closed the connection (browser tab closed)
			return
		}
	}
}
