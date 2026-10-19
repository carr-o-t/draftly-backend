package realtime

import (
	"sync"

	"github.com/google/uuid"
	"github.com/personal-stakshi/draftly/backend/database"
)

type Hub struct {
	rooms map[uuid.UUID]*Room
	mu    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[uuid.UUID]*Room),
	}
}

func (h *Hub) GetOrCreateRoom(id uuid.UUID) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[id]
	if !ok {
		room = NewRoom(id, database.DB)
		h.rooms[id] = room
		go room.Run()
	}
	return room
}

// empty rooms are left running for now, they are cheap (just a goroutine waiting on channel)
