package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/db"
)

type NotificationsHandler struct{}

func NewNotificationsHandler() *NotificationsHandler {
	return &NotificationsHandler{}
}

func (h *NotificationsHandler) ListMine(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id::text, title, body, channel, status, scheduled_for::text, created_at::text
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type row struct {
		ID           string  `json:"id"`
		Title        string  `json:"title"`
		Body         string  `json:"body"`
		Channel      string  `json:"channel"`
		Status       string  `json:"status"`
		ScheduledFor *string `json:"scheduled_for"`
		CreatedAt    string  `json:"created_at"`
	}
	var out []row
	for rows.Next() {
		var r row
		var sched sql.NullString
		if err := rows.Scan(&r.ID, &r.Title, &r.Body, &r.Channel, &r.Status, &sched, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		if sched.Valid {
			s := sched.String
			r.ScheduledFor = &s
		}
		out = append(out, r)
	}
	if out == nil {
		out = []row{}
	}
	c.JSON(http.StatusOK, gin.H{"notifications": out})
}
