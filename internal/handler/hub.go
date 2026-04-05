package handler

import (
	"sync"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

type Hub struct {
	clients map[chan domain.ProgressEvent]struct{}
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[chan domain.ProgressEvent]struct{}),
	}
}

func (h *Hub) Register(client chan domain.ProgressEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client] = struct{}{}
}

func (h *Hub) Unregister(client chan domain.ProgressEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, client)
}

func (h *Hub) Broadcast(event domain.ProgressEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		select {
		case client <- event:
		default:
			// client is slow/full — skip, don't block
		}
	}
}
