package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type wsClient struct {
	hub    *MessagingHub
	userID uuid.UUID
	conn   *websocket.Conn
	send   chan []byte
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // browsers send Origin; tighten alongside API CORS in production
	},
}

func (c *wsClient) readPump() {
	defer func() {
		if c.hub != nil {
			c.hub.unregister <- c
		}
		_ = c.conn.Close()
	}()
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *wsClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWebSocket upgrades GET /ws with query token=<JWT>.
func (mh *MessagingHandler) ServeWebSocket(auth *AuthHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if mh.hub == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Realtime messaging unavailable"})
			return
		}
		token := strings.TrimSpace(c.Query("token"))
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token query parameter required"})
			return
		}
		claims, err := auth.ParseJWTClaims(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		if claims.Role != "patient" && claims.Role != "doctor" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client := &wsClient{
			hub:    mh.hub,
			userID: claims.UserID,
			conn:   conn,
			send:   make(chan []byte, 256),
		}
		mh.hub.register <- client
		go client.writePump()
		go client.readPump()
	}
}
