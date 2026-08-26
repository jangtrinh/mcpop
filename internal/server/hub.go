package server

import (
	"encoding/json"
	"fmt"
	"sync"
)

type SSEHub struct {
	clients map[chan []byte]bool
	mu      sync.RWMutex
}

func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[chan []byte]bool),
	}
}

func (h *SSEHub) Register(ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[ch] = true
}

func (h *SSEHub) Unregister(ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[ch]; ok {
		delete(h.clients, ch)
		close(ch)
	}
}

func (h *SSEHub) Broadcast(eventType string, data interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(jsonData))
	rawBytes := []byte(msg)

	for ch := range h.clients {
		select {
		case ch <- rawBytes:
		default:
			// Client buffer full, skip
		}
	}
}
