package realtime

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// how long we wait before writing a message to the peer
	writeWait = 10 * time.Second

	// how long we wait for the next pong from the peer
	pongWait = 60 * time.Second

	// how often we ping the peer, must be less than pongwait
	pingPeriod = (pongWait * 9) / 10

	// biggest message we accept from the peer
	maxMessageSize = 1024 * 1024 * 5
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// allow all origins for now, tighten this in production
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// client sits between the websocket connection and the hub
type Client struct {
	hub *Hub

	// the websocket connection
	conn *websocket.Conn

	// buffered channle of outbound messages
	send chan []byte

	// the room this client belongs to
	room *Room

	// viewers can receive but never send edits
	ReadOnly bool
}

// readPump reads messages from the websocket and sends them to the hub
func (c *Client) readPump() {
	defer func() {
		c.room.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		// yjs uses binary messages
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if messageType == websocket.BinaryMessage {
			// viewers can send cursor presence and the sync handshake,
			// they just can't send document edits. block only the sync
			// update message, which carries the actual doc change.
			if c.ReadOnly && syncMsgUpdate(message) {
				continue
			}
			c.room.broadcast <- BroadcastMessage{
				SourceClient: c,
				Data:         message,
			}
		}
	}
}

// writePump writes messages from the hub to the websocket
// each yjs message goes as its own frame, the client only reads
// one message per frame
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// the hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
