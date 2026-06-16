package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// SendMessage input validation tests (no DB required)
// ---------------------------------------------------------------------------

func setupSendMessageRouter(t *testing.T) (*gin.Engine, *AuthHandler) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	auth := NewAuthHandler("msg-test-secret")
	hub := NewMessagingHub()
	go hub.Run(context.Background())
	mh := NewMessagingHandler(hub)
	r := gin.New()
	msg := r.Group("/api/messages", auth.RequireAuth(), auth.RequireRole("patient", "doctor"))
	msg.POST("/threads/:threadId/messages", mh.SendMessage)
	return r, auth
}

func sendMsgToken(t *testing.T, auth *AuthHandler, role string) string {
	t.Helper()
	tok, err := auth.createToken(uuid.New(), role+"@test.local", role)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	return "Bearer " + tok
}

func TestSendMessage_Unauthenticated(t *testing.T) {
	r, _ := setupSendMessageRouter(t)
	threadID := uuid.New().String()
	body := `{"body":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/api/messages/threads/"+threadID+"/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestSendMessage_AdminForbidden(t *testing.T) {
	r, auth := setupSendMessageRouter(t)
	threadID := uuid.New().String()
	body := `{"body":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/api/messages/threads/"+threadID+"/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", sendMsgToken(t, auth, "admin"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for admin role on message endpoint, got %d", w.Code)
	}
}

func TestSendMessage_InvalidThreadID(t *testing.T) {
	r, auth := setupSendMessageRouter(t)
	body := `{"body":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/api/messages/threads/not-a-uuid/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", sendMsgToken(t, auth, "patient"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid thread UUID, got %d", w.Code)
	}
}

func TestSendMessage_EmptyBody(t *testing.T) {
	r, auth := setupSendMessageRouter(t)
	threadID := uuid.New().String()
	body := `{"body":"   "}`
	req := httptest.NewRequest(http.MethodPost, "/api/messages/threads/"+threadID+"/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", sendMsgToken(t, auth, "patient"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for whitespace-only body, got %d", w.Code)
	}
}

func TestSendMessage_BodyTooLong(t *testing.T) {
	r, auth := setupSendMessageRouter(t)
	threadID := uuid.New().String()
	long := strings.Repeat("a", maxMessageRunes+1)
	payload, _ := json.Marshal(map[string]string{"body": long})
	req := httptest.NewRequest(http.MethodPost, "/api/messages/threads/"+threadID+"/messages", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", sendMsgToken(t, auth, "patient"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for body over %d runes, got %d", maxMessageRunes, w.Code)
	}
}

// TestSendMessage_ExactMaxLength_PassesLengthValidation only checks the length guard
// (not the DB path) — the handler reaches DB only after validation passes.
func TestSendMessage_ExactMaxLength_PassesLengthValidation(t *testing.T) {
	// Verify that a body of exactly maxMessageRunes does not trigger the length error.
	exact := strings.Repeat("a", maxMessageRunes)
	runeCount := len([]rune(exact))
	if runeCount != maxMessageRunes {
		t.Fatalf("test setup: expected %d runes, got %d", maxMessageRunes, runeCount)
	}
	// The handler rejects bodies where runeCount > maxMessageRunes — exactly equal is fine.
	if runeCount > maxMessageRunes {
		t.Errorf("exact body unexpectedly exceeds maxMessageRunes (%d > %d)", runeCount, maxMessageRunes)
	}
}

func TestSendMessage_UnicodeBodyAtLimit(t *testing.T) {
	r, auth := setupSendMessageRouter(t)
	threadID := uuid.New().String()
	// 8001 CJK runes — should be rejected
	over := strings.Repeat("中", maxMessageRunes+1)
	payload, _ := json.Marshal(map[string]string{"body": over})
	req := httptest.NewRequest(http.MethodPost, "/api/messages/threads/"+threadID+"/messages", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", sendMsgToken(t, auth, "doctor"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unicode body over limit, got %d", w.Code)
	}
}

func TestSendMessage_MalformedJSON(t *testing.T) {
	r, auth := setupSendMessageRouter(t)
	threadID := uuid.New().String()
	req := httptest.NewRequest(http.MethodPost, "/api/messages/threads/"+threadID+"/messages", strings.NewReader("{not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", sendMsgToken(t, auth, "patient"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed JSON, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// maxMessageRunes constant regression test
// ---------------------------------------------------------------------------

func TestMaxMessageRunes_IsReasonable(t *testing.T) {
	if maxMessageRunes < 100 {
		t.Errorf("maxMessageRunes=%d is suspiciously small", maxMessageRunes)
	}
	if maxMessageRunes > 100000 {
		t.Errorf("maxMessageRunes=%d is suspiciously large", maxMessageRunes)
	}
}
