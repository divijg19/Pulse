package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/divijg19/Pulse/internal/model"
	"github.com/divijg19/Pulse/internal/runconfig"
	"github.com/divijg19/Pulse/internal/stream"
)

type StreamHandler struct {
	Hub *stream.Hub
}

// streamClientBuffer sizes each SSE subscriber's event channel. A single run
// emits at most runconfig.MaxConcurrency results, so this multiple leaves
// headroom for overlapping runs and several concurrent subscribers before the
// Hub's drop-on-full policy begins discarding events.
const streamClientBuffer = runconfig.MaxConcurrency * 8

func (h *StreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Buffer well beyond a single run's event volume so overlapping runs and
	// multiple subscribers don't immediately overflow and lose results. The
	// Hub still drops on a full buffer to keep the engine from stalling, so
	// this only widens the window before a slow client starts missing events.
	clientChan := make(chan model.Event, streamClientBuffer)
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
