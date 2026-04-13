package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/db"
)

type AdminAuditHandler struct{}

func NewAdminAuditHandler() *AdminAuditHandler {
	return &AdminAuditHandler{}
}

func (h *AdminAuditHandler) List(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id::text, actor_user_id::text, action, entity_type, entity_id, detail, created_at::text
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT 200
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	type row struct {
		ID          string `json:"id"`
		ActorUserID string `json:"actor_user_id"`
		Action      string `json:"action"`
		EntityType  string `json:"entity_type"`
		EntityID    string `json:"entity_id"`
		Detail      string `json:"detail"`
		CreatedAt   string `json:"created_at"`
	}
	var out []row
	for rows.Next() {
		var r row
		var actor sql.NullString
		if err := rows.Scan(&r.ID, &actor, &r.Action, &r.EntityType, &r.EntityID, &r.Detail, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		if actor.Valid {
			r.ActorUserID = actor.String
		}
		out = append(out, r)
	}
	if out == nil {
		out = []row{}
	}
	c.JSON(http.StatusOK, gin.H{"entries": out})
}
