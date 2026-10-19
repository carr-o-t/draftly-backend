package realtime

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/database"
	"github.com/personal-stakshi/draftly/backend/utils"
)

// bearerToken pulls the token out of an "Authorization: Bearer <token>" header.
func bearerToken(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

// ServeWs handles a websocket request for a document room.
func ServeWs(hub *Hub, c *gin.Context) {
	idStr := c.Param("id")
	docID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	// check auth before upgrading so we can send a normal http error
	token := c.Query("token")
	if token == "" {
		token = bearerToken(c.GetHeader("Authorization"))
	}

	cfg := config.LoadConfig()
	claims, tokenErr := utils.ValidateToken(token, cfg)
	if tokenErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	role, roleErr := resolveRole(database.DB, docID, claims.UserID)
	if roleErr != nil {
		if roleErr.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		}
		return
	}

	// upgrade the http connection to websocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	room := hub.GetOrCreateRoom(docID)

	client := &Client{
		hub:      hub,
		room:     room,
		conn:     conn,
		send:     make(chan []byte, 256),
		ReadOnly: role == "viewer",
	}
	client.room.register <- client

	// do the rest of the work in background goroutines
	go client.writePump()
	go client.readPump()
}
