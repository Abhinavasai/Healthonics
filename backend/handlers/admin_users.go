package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
)

type AdminUsersHandler struct{}

func NewAdminUsersHandler() *AdminUsersHandler {
	return &AdminUsersHandler{}
}

type adminUserRow struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Active bool   `json:"active"`
}

// ListUsers returns all users for admin management.
func (h *AdminUsersHandler) ListUsers(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok || claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id, email, role, COALESCE(active, TRUE) FROM users ORDER BY role, email
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	out := make([]adminUserRow, 0)
	for rows.Next() {
		var r adminUserRow
		var id uuid.UUID
		if err := rows.Scan(&id, &r.Email, &r.Role, &r.Active); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		r.ID = id.String()
		out = append(out, r)
	}
	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": out})
}

type patchUserRequest struct {
	Active *bool `json:"active"`
}

// PatchUser sets active flag (admin cannot disable self).
func (h *AdminUsersHandler) PatchUser(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok || claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}
	if id == claims.UserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot change your own account this way"})
		return
	}

	var req patchUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if req.Active == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "active is required"})
		return
	}

	cmd, err := db.Pool.Exec(c.Request.Context(), `
		UPDATE users SET active = $1 WHERE id = $2
	`, *req.Active, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if cmd.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
