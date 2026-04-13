package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5"
)

type MessagingHandler struct{}

func NewMessagingHandler() *MessagingHandler {
	return &MessagingHandler{}
}

const maxMessageRunes = 8000

func previewText(s string) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= 160 {
		return s
	}
	runes := []rune(s)
	return string(runes[:160]) + "…"
}

// validPatientDoctorPair returns (patientID, doctorID) if peer is the opposite role of caller.
func validPatientDoctorPair(meRole string, meID uuid.UUID, peerID uuid.UUID, peerRole string) (patientID, doctorID uuid.UUID, ok bool) {
	switch meRole {
	case "patient":
		if peerRole != "doctor" {
			return uuid.Nil, uuid.Nil, false
		}
		return meID, peerID, true
	case "doctor":
		if peerRole != "patient" {
			return uuid.Nil, uuid.Nil, false
		}
		return peerID, meID, true
	default:
		return uuid.Nil, uuid.Nil, false
	}
}

type createThreadRequest struct {
	PeerUserID string `json:"peer_user_id" binding:"required"`
	Body       string `json:"body" binding:"required"`
}

type sendMessageRequest struct {
	Body string `json:"body" binding:"required"`
}

type messageThreadRow struct {
	ID            uuid.UUID `json:"id"`
	PeerUserID    uuid.UUID `json:"peer_user_id"`
	PeerEmail     string    `json:"peer_email"`
	LastMessageAt time.Time `json:"last_message_at"`
	LastPreview   string    `json:"last_preview"`
	UnreadCount   int       `json:"unread_count"`
}

type chatMessageRow struct {
	ID        uuid.UUID `json:"id"`
	ThreadID  uuid.UUID `json:"thread_id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// UnreadTotal implements GET /api/messages/unread
func (h *MessagingHandler) UnreadTotal(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	var total int
	err := db.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(sub.c), 0)::int
		FROM (
			SELECT (
				SELECT COUNT(*)::int
				FROM messages m
				WHERE m.thread_id = mt.id
				  AND m.sender_id <> $1
				  AND m.created_at > COALESCE(
					(SELECT r.last_read_at FROM message_thread_reads r
					 WHERE r.thread_id = mt.id AND r.user_id = $1),
					'-infinity'::timestamptz
				  )
			) AS c
			FROM message_threads mt
			WHERE mt.patient_id = $1 OR mt.doctor_id = $1
		) sub
	`, claims.UserID).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load unread count"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"unread_total": total})
}

// ListThreads implements GET /api/messages/threads
func (h *MessagingHandler) ListThreads(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	rows, err := db.Pool.Query(ctx, `
		SELECT
			t.id,
			CASE WHEN $1 = t.patient_id THEN t.doctor_id ELSE t.patient_id END,
			u.email,
			t.last_message_at,
			t.last_preview,
			(
				SELECT COUNT(*)::int
				FROM messages m
				WHERE m.thread_id = t.id
				  AND m.sender_id <> $1
				  AND m.created_at > COALESCE(
					(SELECT r.last_read_at FROM message_thread_reads r
					 WHERE r.thread_id = t.id AND r.user_id = $1),
					'-infinity'::timestamptz
				  )
			)
		FROM message_threads t
		JOIN users u ON u.id = (CASE WHEN $1 = t.patient_id THEN t.doctor_id ELSE t.patient_id END)
		WHERE t.patient_id = $1 OR t.doctor_id = $1
		ORDER BY t.last_message_at DESC
	`, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load threads"})
		return
	}
	defer rows.Close()

	var out []messageThreadRow
	for rows.Next() {
		var r messageThreadRow
		if err := rows.Scan(&r.ID, &r.PeerUserID, &r.PeerEmail, &r.LastMessageAt, &r.LastPreview, &r.UnreadCount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load threads"})
			return
		}
		out = append(out, r)
	}
	if out == nil {
		out = []messageThreadRow{}
	}
	c.JSON(http.StatusOK, gin.H{"threads": out})
}

// ListMessages implements GET /api/messages/threads/:threadId
func (h *MessagingHandler) ListMessages(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	threadID, err := uuid.Parse(c.Param("threadId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread id"})
		return
	}
	ctx := c.Request.Context()
	if !participantInThread(ctx, threadID, claims.UserID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	rows, err := db.Pool.Query(ctx, `
		SELECT id, thread_id, sender_id, body, created_at
		FROM messages
		WHERE thread_id = $1
		ORDER BY created_at ASC
	`, threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load messages"})
		return
	}
	defer rows.Close()

	var list []chatMessageRow
	for rows.Next() {
		var m chatMessageRow
		if err := rows.Scan(&m.ID, &m.ThreadID, &m.SenderID, &m.Body, &m.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load messages"})
			return
		}
		list = append(list, m)
	}
	if list == nil {
		list = []chatMessageRow{}
	}
	c.JSON(http.StatusOK, gin.H{"messages": list})
}

// SendMessage implements POST /api/messages/threads/:threadId/messages
func (h *MessagingHandler) SendMessage(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	threadID, err := uuid.Parse(c.Param("threadId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread id"})
		return
	}
	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	body := strings.TrimSpace(req.Body)
	if body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is required"})
		return
	}
	if utf8.RuneCountInString(body) > maxMessageRunes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message too long"})
		return
	}
	ctx := c.Request.Context()
	if !participantInThread(ctx, threadID, claims.UserID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	var msg chatMessageRow
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO messages (thread_id, sender_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, thread_id, sender_id, body, created_at
	`, threadID, claims.UserID, body).Scan(&msg.ID, &msg.ThreadID, &msg.SenderID, &msg.Body, &msg.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to send message"})
		return
	}
	_, _ = db.Pool.Exec(ctx, `
		UPDATE message_threads
		SET last_message_at = $2, last_preview = $3
		WHERE id = $1
	`, threadID, msg.CreatedAt, previewText(body))

	c.JSON(http.StatusCreated, msg)
}

// CreateThread implements POST /api/messages/threads
func (h *MessagingHandler) CreateThread(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	var req createThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	peerID, err := uuid.Parse(strings.TrimSpace(req.PeerUserID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "peer_user_id must be a valid UUID"})
		return
	}
	body := strings.TrimSpace(req.Body)
	if body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is required"})
		return
	}
	if utf8.RuneCountInString(body) > maxMessageRunes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message too long"})
		return
	}
	if peerID == claims.UserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot message yourself"})
		return
	}

	ctx := c.Request.Context()
	var peerRole string
	err = db.Pool.QueryRow(ctx, `SELECT role FROM users WHERE id = $1`, peerID).Scan(&peerRole)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Peer user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	patientID, doctorID, pairOK := validPatientDoctorPair(claims.Role, claims.UserID, peerID, peerRole)
	if !pairOK {
		c.JSON(http.StatusForbidden, gin.H{"error": "Messaging is only allowed between a patient and a doctor"})
		return
	}

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var threadID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT id FROM message_threads WHERE patient_id = $1 AND doctor_id = $2
	`, patientID, doctorID).Scan(&threadID)
	if err == pgx.ErrNoRows {
		err = tx.QueryRow(ctx, `
			INSERT INTO message_threads (patient_id, doctor_id, last_message_at, last_preview)
			VALUES ($1, $2, NOW(), $3)
			RETURNING id
		`, patientID, doctorID, previewText(body)).Scan(&threadID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create thread"})
		return
	}

	var msgTime time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO messages (thread_id, sender_id, body)
		VALUES ($1, $2, $3)
		RETURNING created_at
	`, threadID, claims.UserID, body).Scan(&msgTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to send first message"})
		return
	}

	_, err = tx.Exec(ctx, `
		UPDATE message_threads SET last_message_at = $2, last_preview = $3 WHERE id = $1
	`, threadID, msgTime, previewText(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update thread"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	var row messageThreadRow
	err = db.Pool.QueryRow(ctx, `
		SELECT
			t.id,
			CASE WHEN $1 = t.patient_id THEN t.doctor_id ELSE t.patient_id END,
			u.email,
			t.last_message_at,
			t.last_preview,
			(
				SELECT COUNT(*)::int
				FROM messages m
				WHERE m.thread_id = t.id
				  AND m.sender_id <> $1
				  AND m.created_at > COALESCE(
					(SELECT r.last_read_at FROM message_thread_reads r
					 WHERE r.thread_id = t.id AND r.user_id = $1),
					'-infinity'::timestamptz
				  )
			)
		FROM message_threads t
		JOIN users u ON u.id = (CASE WHEN $1 = t.patient_id THEN t.doctor_id ELSE t.patient_id END)
		WHERE t.id = $2
	`, claims.UserID, threadID).Scan(&row.ID, &row.PeerUserID, &row.PeerEmail, &row.LastMessageAt, &row.LastPreview, &row.UnreadCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load thread"})
		return
	}

	c.JSON(http.StatusCreated, row)
}

// MarkRead implements POST /api/messages/threads/:threadId/read
func (h *MessagingHandler) MarkRead(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	threadID, err := uuid.Parse(c.Param("threadId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread id"})
		return
	}
	ctx := c.Request.Context()
	if !participantInThread(ctx, threadID, claims.UserID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO message_thread_reads (thread_id, user_id, last_read_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (thread_id, user_id) DO UPDATE SET last_read_at = EXCLUDED.last_read_at
	`, threadID, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to mark read"})
		return
	}
	c.Status(http.StatusNoContent)
}

func participantInThread(ctx context.Context, threadID, userID uuid.UUID) bool {
	var n int
	err := db.Pool.QueryRow(ctx, `
		SELECT 1 FROM message_threads
		WHERE id = $1 AND (patient_id = $2 OR doctor_id = $2)
	`, threadID, userID).Scan(&n)
	return err == nil
}
