package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type AdminUserLifecycleHandler struct{}

type updateLifecycleSettingsRequest struct {
	NewUserWindowDays  int    `json:"new_user_window_days"`
	InactiveWindowDays int    `json:"inactive_window_days"`
	Reason             string `json:"reason"`
	Confirm            string `json:"confirm"`
}

type createLifecycleUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type updateLifecycleUserRequest struct {
	Email *string `json:"email"`
	Role  *string `json:"role"`
}

type resetLifecyclePasswordRequest struct {
	NewPassword string `json:"new_password"`
	Reason      string `json:"reason"`
	Confirm     string `json:"confirm"`
}

type deactivateLifecycleUserRequest struct {
	Reason  string `json:"reason"`
	Confirm string `json:"confirm"`
}

func NewAdminUserLifecycleHandler() *AdminUserLifecycleHandler {
	return &AdminUserLifecycleHandler{}
}

func validLifecycleRole(role string) bool {
	switch role {
	case "patient", "doctor", "admin":
		return true
	default:
		return false
	}
}

func validLifecyclePassword(pw string) bool {
	return utf8.RuneCountInString(strings.TrimSpace(pw)) >= 6
}

func (h *AdminUserLifecycleHandler) ListUsers(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	rows, err := db.Pool.Query(c.Request.Context(), `
		SELECT id::text, email, role, is_active, created_at::text, COALESCE(deactivated_at::text, ''), COALESCE(deactivated_by::text, '')
		FROM users
		ORDER BY created_at DESC
		LIMIT 500
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()
	type row struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Role          string `json:"role"`
		IsActive      bool   `json:"is_active"`
		CreatedAt     string `json:"created_at"`
		DeactivatedAt string `json:"deactivated_at"`
		DeactivatedBy string `json:"deactivated_by"`
	}
	out := make([]row, 0, 32)
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ID, &r.Email, &r.Role, &r.IsActive, &r.CreatedAt, &r.DeactivatedAt, &r.DeactivatedBy); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": out})
}

func (h *AdminUserLifecycleHandler) GetSettingsKpis(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}

	type settingsRow struct {
		NewUserWindowDays  int    `json:"new_user_window_days"`
		InactiveWindowDays int    `json:"inactive_window_days"`
		UpdatedAt          string `json:"updated_at"`
		UpdatedBy          string `json:"updated_by"`
	}
	type kpisRow struct {
		TotalUsers         int `json:"total_users"`
		TotalPatients      int `json:"total_patients"`
		TotalDoctors       int `json:"total_doctors"`
		TotalAdmins        int `json:"total_admins"`
		NewUsersInWindow   int `json:"new_users_in_window"`
		InactiveUsersCount int `json:"inactive_users_count"`
	}

	var settings settingsRow
	if err := db.Pool.QueryRow(c.Request.Context(), `
		SELECT
			new_user_window_days,
			inactive_window_days,
			updated_at::text,
			COALESCE(updated_by::text, '')
		FROM admin_user_lifecycle_settings
		WHERE id = TRUE
	`).Scan(&settings.NewUserWindowDays, &settings.InactiveWindowDays, &settings.UpdatedAt, &settings.UpdatedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	var kpis kpisRow
	if err := db.Pool.QueryRow(c.Request.Context(), `
		WITH cfg AS (
			SELECT new_user_window_days, inactive_window_days
			FROM admin_user_lifecycle_settings
			WHERE id = TRUE
		),
		last_activity AS (
			SELECT user_id, MAX(event_at) AS last_event_at
			FROM (
				SELECT id AS user_id, created_at AS event_at FROM users
				UNION ALL
				SELECT patient_id AS user_id, created_at AS event_at FROM appointments
				UNION ALL
				SELECT doctor_id AS user_id, created_at AS event_at FROM appointments
				UNION ALL
				SELECT sender_id AS user_id, created_at AS event_at FROM messages
			) events
			GROUP BY user_id
		)
		SELECT
			COUNT(*)::int AS total_users,
			COUNT(*) FILTER (WHERE u.role = 'patient')::int AS total_patients,
			COUNT(*) FILTER (WHERE u.role = 'doctor')::int AS total_doctors,
			COUNT(*) FILTER (WHERE u.role = 'admin')::int AS total_admins,
			COUNT(*) FILTER (
				WHERE u.created_at >= NOW() - make_interval(days => cfg.new_user_window_days)
			)::int AS new_users_in_window,
			COUNT(*) FILTER (
				WHERE COALESCE(la.last_event_at, u.created_at) < NOW() - make_interval(days => cfg.inactive_window_days)
			)::int AS inactive_users_count
		FROM users u
		CROSS JOIN cfg
		LEFT JOIN last_activity la ON la.user_id = u.id
	`).Scan(
		&kpis.TotalUsers,
		&kpis.TotalPatients,
		&kpis.TotalDoctors,
		&kpis.TotalAdmins,
		&kpis.NewUsersInWindow,
		&kpis.InactiveUsersCount,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
		"kpis":     kpis,
	})
}

func (h *AdminUserLifecycleHandler) UpdateSettings(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	var req updateLifecycleSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if req.NewUserWindowDays < 1 || req.NewUserWindowDays > 3650 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new_user_window_days must be between 1 and 3650"})
		return
	}
	if req.InactiveWindowDays < 1 || req.InactiveWindowDays > 3650 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "inactive_window_days must be between 1 and 3650"})
		return
	}
	if err := validateHighRiskGuardrail(req.Reason, req.Confirm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tx, err := db.Pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()
	if _, err := tx.Exec(c.Request.Context(), `
		UPDATE admin_user_lifecycle_settings
		SET
			new_user_window_days = $1,
			inactive_window_days = $2,
			updated_by = $3,
			updated_at = NOW()
		WHERE id = TRUE
	`, req.NewUserWindowDays, req.InactiveWindowDays, claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	detail := fmt.Sprintf("updated lifecycle settings new_user_window_days=%d inactive_window_days=%d", req.NewUserWindowDays, req.InactiveWindowDays)
	detail = appendAuditReason(detail, req.Reason)
	if err := writeAudit(c.Request.Context(), tx, claims.UserID, "admin_user_lifecycle_settings_updated", "admin_user_lifecycle_settings", "singleton", detail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	h.GetSettingsKpis(c)
}

func (h *AdminUserLifecycleHandler) CreateUser(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	var req createLifecycleUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Role = strings.TrimSpace(strings.ToLower(req.Role))
	if !validEmailFormat(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid email is required"})
		return
	}
	if !validLifecycleRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be patient, doctor, or admin"})
		return
	}
	if !validLifecyclePassword(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 6 characters"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	tx, err := db.Pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()

	var id string
	err = tx.QueryRow(c.Request.Context(), `
		INSERT INTO users(email, password_hash, role, is_active)
		VALUES ($1, $2, $3, TRUE)
		RETURNING id::text
	`, req.Email, string(hash), req.Role).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	detail := fmt.Sprintf("created user email=%s role=%s", req.Email, req.Role)
	if err := writeAudit(c.Request.Context(), tx, claims.UserID, "admin_user_created", "user", id, detail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "email": req.Email, "role": req.Role, "is_active": true})
}

func (h *AdminUserLifecycleHandler) UpdateUser(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	userID := strings.TrimSpace(c.Param("id"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}
	var req updateLifecycleUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if req.Email == nil && req.Role == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field is required"})
		return
	}

	tx, err := db.Pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()

	var oldEmail, oldRole string
	var isActive bool
	err = tx.QueryRow(c.Request.Context(), `SELECT email, role, is_active FROM users WHERE id::text = $1`, userID).Scan(&oldEmail, &oldRole, &isActive)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	nextEmail := oldEmail
	nextRole := oldRole
	if req.Email != nil {
		nextEmail = strings.TrimSpace(strings.ToLower(*req.Email))
		if !validEmailFormat(nextEmail) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Valid email is required"})
			return
		}
	}
	if req.Role != nil {
		nextRole = strings.TrimSpace(strings.ToLower(*req.Role))
		if !validLifecycleRole(nextRole) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "role must be patient, doctor, or admin"})
			return
		}
	}
	if _, err := tx.Exec(c.Request.Context(), `
		UPDATE users
		SET email = $2,
		    role = $3
		WHERE id::text = $1
	`, userID, nextEmail, nextRole); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	detail := fmt.Sprintf("updated user email:%s->%s role:%s->%s", oldEmail, nextEmail, oldRole, nextRole)
	if err := writeAudit(c.Request.Context(), tx, claims.UserID, "admin_user_updated", "user", userID, detail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": userID, "email": nextEmail, "role": nextRole, "is_active": isActive})
}

func (h *AdminUserLifecycleHandler) DeactivateUser(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	userID := strings.TrimSpace(c.Param("id"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}
	if claims.UserID.String() == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot deactivate current admin"})
		return
	}
	var req deactivateLifecycleUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := validateHighRiskGuardrail(req.Reason, req.Confirm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.Pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()

	var email string
	err = tx.QueryRow(c.Request.Context(), `
		UPDATE users
		SET is_active = FALSE,
		    deactivated_at = NOW(),
		    deactivated_by = $2
		WHERE id::text = $1
		  AND is_active = TRUE
		RETURNING email
	`, userID, claims.UserID).Scan(&email)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Active user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	detail := appendAuditReason("deactivated user "+email, req.Reason)
	if err := writeAudit(c.Request.Context(), tx, claims.UserID, "admin_user_deactivated", "user", userID, detail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": userID, "status": "deactivated"})
}

func (h *AdminUserLifecycleHandler) ResetPassword(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	userID := strings.TrimSpace(c.Param("id"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}
	var req resetLifecyclePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if !validLifecyclePassword(req.NewPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new_password must be at least 6 characters"})
		return
	}
	if err := validateHighRiskGuardrail(req.Reason, req.Confirm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	tx, err := db.Pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()

	cmd, err := tx.Exec(c.Request.Context(), `UPDATE users SET password_hash = $2 WHERE id::text = $1`, userID, string(hash))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if cmd.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	detail := appendAuditReason("password reset by admin", req.Reason)
	if err := writeAudit(c.Request.Context(), tx, claims.UserID, "admin_user_password_reset", "user", userID, detail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": userID, "password_reset": true})
}
