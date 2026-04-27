package handlers

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
)

// MessagingHub tracks authenticated WebSocket connections per user and pushes JSON events (e.g. new_message).
type MessagingHub struct {
	mu         sync.RWMutex
	clients    map[uuid.UUID]map[*wsClient]bool
	register   chan *wsClient
	unregister chan *wsClient
}

func NewMessagingHub() *MessagingHub {
	return &MessagingHub{
		clients:    make(map[uuid.UUID]map[*wsClient]bool),
		register:   make(chan *wsClient),
		unregister: make(chan *wsClient),
	}
}

// Run processes registration lifecycle until Stop is integrated (runs forever).
func (h *MessagingHub) Run() {
	for {
		select {
		case client := <-h.register:
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*wsClient]bool)
			}
			h.clients[client.userID][client] = true
		case client := <-h.unregister:
			if m, ok := h.clients[client.userID]; ok {
				delete(m, client)
				if len(m) == 0 {
					delete(h.clients, client.userID)
				}
			}
		}
	}
}

// NotifyNewMessage notifies both thread participants over WebSocket (best-effort).
func (h *MessagingHub) NotifyNewMessage(ctx context.Context, threadID uuid.UUID, msg chatMessageRow) {
	if h == nil {
		return
	}
	var patientID, doctorID uuid.UUID
	err := db.Pool.QueryRow(ctx, `
		SELECT patient_id, doctor_id FROM message_threads WHERE id = $1
	`, threadID).Scan(&patientID, &doctorID)
	if err != nil {
		return
	}
	payload, err := json.Marshal(map[string]interface{}{
		"type":      "new_message",
		"thread_id": threadID.String(),
		"message":   msg,
	})
	if err != nil {
		return
	}

	h.mu.RLock()
	var targets []*wsClient
	for _, uid := range []uuid.UUID{patientID, doctorID} {
		for c := range h.clients[uid] {
			targets = append(targets, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range targets {
		select {
		case c.send <- payload:
		default:
			// Slow consumer — skip this tick (polling fallback still works).
		}
	}
}
