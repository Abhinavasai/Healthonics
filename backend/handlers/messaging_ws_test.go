package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestServeWebSocket_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := NewAuthHandler("test-secret-key-for-jwt-parse-tests-xx")
	hub := NewMessagingHub()
	mh := NewMessagingHandler(hub)
	h := mh.ServeWebSocket(auth)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/messages/ws", nil)
	h(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestServeWebSocket_HubNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := NewAuthHandler("test-secret-key-for-jwt-parse-tests-xx")
	mh := NewMessagingHandler(nil)
	h := mh.ServeWebSocket(auth)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/messages/ws?token=x", nil)
	h(c)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestServeWebSocket_AdminForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := NewAuthHandler("test-secret-key-for-jwt-parse-tests-xx")
	tok, err := auth.createToken(uuid.New(), "admin@test.local", "admin")
	if err != nil {
		t.Fatal(err)
	}
	hub := NewMessagingHub()
	mh := NewMessagingHandler(hub)
	h := mh.ServeWebSocket(auth)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/messages/ws?token="+tok, nil)
	h(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", w.Code, w.Body.String())
	}
}
