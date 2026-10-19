package realtime

import (
	"log"

	"github.com/google/uuid"
	"github.com/personal-stakshi/draftly/backend/models"
	"gorm.io/gorm"
)

// BroadcastMessage is a payload for one client or everyone
type BroadcastMessage struct {
	SourceClient *Client // origin, skipped during broadcast
	TargetClient *Client // set for unicast, nil for broadcast
	Data         []byte
}

// room keeps active clients and broadcasts messages within a room
type Room struct {
	ID         uuid.UUID
	clients    map[*Client]bool
	broadcast  chan BroadcastMessage
	register   chan *Client
	unregister chan *Client
	db         *gorm.DB
}

func NewRoom(id uuid.UUID, db *gorm.DB) *Room {
	return &Room{
		ID:         id,
		broadcast:  make(chan BroadcastMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		db:         db,
	}
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.register:
			// add client and load history off the event loop
			r.clients[client] = true
			go r.loadHistory(client)

		case client := <-r.unregister:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
			}

		case message := <-r.broadcast:
			if message.TargetClient != nil {
				target := message.TargetClient
				if _, ok := r.clients[target]; ok {
					select {
					case target.send <- message.Data:
					default:
						// drop slow or disconnected client
						close(target.send)
						delete(r.clients, target)
					}
				}
				continue
			}

			// save to db before sending to others
			r.handleMessage(message)

			// send to everyone except the sender
			for client := range r.clients {
				if client == message.SourceClient {
					continue
				}

				select {
				case client.send <- message.Data:
				default:
					// drop slow or disconnected clients
					close(client.send)
					delete(r.clients, client)
				}
			}
		}
	}
}

func (r *Room) handleMessage(msg BroadcastMessage) {
	if len(msg.Data) == 0 {
		return
	}

	// only save real doc updates, not handshake steps
	if syncMsgUpdate(msg.Data) {
		go r.saveUpdate(msg.Data)
	}
	// cursor and awareness messages are sent but not saved
}

func (r *Room) saveUpdate(data []byte) {
	update := models.DocumentUpdate{
		DocumentID: r.ID,
		Data:       data,
	}
	if err := r.db.Create(&update).Error; err != nil {
		log.Printf("Error saving update: %v", err)
	}
}

func (r *Room) loadHistory(client *Client) {
	var updates []models.DocumentUpdate
	if err := r.db.Where("document_id = ?", r.ID).Order("created_at asc").Find(&updates).Error; err != nil {
		log.Printf("Error loading history: %v", err)
		return
	}

	for _, u := range updates {
		r.broadcast <- BroadcastMessage{TargetClient: client, Data: u.Data}
	}
}
