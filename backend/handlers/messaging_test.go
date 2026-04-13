package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestPreviewText_ShortUnchanged(t *testing.T) {
	s := "hello"
	if previewText(s) != "hello" {
		t.Fatalf("expected short text unchanged")
	}
}

func TestPreviewText_TruncatesLong(t *testing.T) {
	s := strings.Repeat("あ", 200)
	out := previewText(s)
	if utf8.RuneCountInString(out) < 161 {
		t.Fatalf("expected truncation with ellipsis, got rune count %d", utf8.RuneCountInString(out))
	}
	if !strings.HasSuffix(out, "…") {
		t.Fatalf("expected ellipsis suffix")
	}
}

func TestValidPatientDoctorPair(t *testing.T) {
	p := uuid.New()
	d := uuid.New()
	pa, doc, ok := validPatientDoctorPair("patient", p, d, "doctor")
	if !ok || pa != p || doc != d {
		t.Fatalf("expected patient-doctor pair")
	}
	pa, doc, ok = validPatientDoctorPair("doctor", d, p, "patient")
	if !ok || pa != p || doc != d {
		t.Fatalf("expected doctor-patient pair")
	}
	_, _, ok = validPatientDoctorPair("patient", p, d, "patient")
	if ok {
		t.Fatalf("expected invalid when peer is not doctor")
	}
}

func TestUnreadTotal_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/messages/unread", nil)

	h := NewMessagingHandler()
	h.UnreadTotal(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestListMessages_InvalidThreadUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/messages/threads/not-a-uuid", nil)
	c.Params = gin.Params{{Key: "threadId", Value: "not-a-uuid"}}
	c.Set("claims", &Claims{
		UserID: uuid.New(),
		Email:  "a@b.com",
		Role:   "patient",
	})

	h := NewMessagingHandler()
	h.ListMessages(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestSendMessage_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/messages/threads/"+uuid.New().String()+"/messages", bytes.NewReader([]byte(`{`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "threadId", Value: uuid.New().String()}}
	c.Set("claims", &Claims{
		UserID: uuid.New(),
		Email:  "a@b.com",
		Role:   "patient",
	})

	h := NewMessagingHandler()
	h.SendMessage(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateThread_InvalidPeerUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"peer_user_id":"x","body":"hello"}`
	c.Request = httptest.NewRequest("POST", "/api/messages/threads", bytes.NewReader([]byte(body)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("claims", &Claims{
		UserID: uuid.New(),
		Email:  "a@b.com",
		Role:   "patient",
	})

	h := NewMessagingHandler()
	h.CreateThread(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
