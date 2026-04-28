package handlers

import (
	"net"
	"net/http"
	"net/url"
	"os"
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
		return isAllowedWSOrigin(r.Header.Get("Origin"))
	},
}

func normalizeOrigin(raw string) (string, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", false
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", false
	}
	host := strings.ToLower(u.Hostname())
	port := u.Port()
	if port == "" {
		if strings.EqualFold(u.Scheme, "https") {
			port = "443"
		} else if strings.EqualFold(u.Scheme, "http") {
			port = "80"
		}
	}
	return strings.ToLower(u.Scheme) + "://" + net.JoinHostPort(host, port), true
}

func isAllowedWSOrigin(origin string) bool {
	allowedRaw := strings.TrimSpace(os.Getenv("WS_ALLOWED_ORIGINS"))
	if allowedRaw == "" {
		// Keep localhost/dev defaults when no explicit policy is configured.
		allowedRaw = "http://127.0.0.1:4300,http://localhost:4300"
	}
	normalizedOrigin, ok := normalizeOrigin(origin)
	if !ok {
		return false
	}
	for _, item := range strings.Split(allowedRaw, ",") {
		candidate, ok := normalizeOrigin(item)
		if ok && candidate == normalizedOrigin {
			return true
		}
	}
	return false
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
		if !isAllowedWSOrigin(c.GetHeader("Origin")) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Origin not allowed"})
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
