package stream

import (
	"sync"

	"github.com/divijg19/Pulse/internal/model"
)

type Hub struct {
	mu      sync.RWMutex
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
	// A read lock lets concurrent broadcasts fan out in parallel while keeping
	// Add/Remove (write lock) mutually exclusive, so a channel is never closed
	// while a broadcast is sending to it. The sends are non-blocking, so the
	// lock is held only for the brief fan-out, not across any slow client.
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- event:
		default:
			// If the channel is blocked, skip it to avoid blocking the
			// broadcast. Slow clients therefore miss events rather than
			// stalling the engine.
		}
	}
}
