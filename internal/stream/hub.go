package stream

import (
	"sync"

	"github.com/divijg19/Pulse/internal/model"
)

type Hub struct {
	mu      sync.Mutex
	clients map[chan model.Event]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[chan model.Event]bool),
	}
}

func (h *Hub) Add(ch chan model.Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[ch] = true
}

func (h *Hub) Remove(ch chan model.Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[ch]; !ok {
		return
	}
	delete(h.clients, ch)
	close(ch)
}

func (h *Hub) Broadcast(event model.Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for ch := range h.clients {
		select {
		case ch <- event:
		default:
			// If the channel is blocked, skip it to avoid blocking the entire broadcast
		}
		// drop slow clients to prevent blocking the hub
	}
}
