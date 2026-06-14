package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func setupAssistantRouter(t *testing.T) (*gin.Engine, *AuthHandler, *AssistantChatHandler) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	auth := NewAuthHandler("test-secret")
	// No Azure config — will always use fallback.
	assistant := NewAssistantChatHandler("", "", "", "")
	r := gin.New()
	r.POST("/api/assistant/chat", auth.RequireAuth(), assistant.Chat)
	return r, auth, assistant
}

func assistantToken(t *testing.T, auth *AuthHandler, role string) string {
	t.Helper()
	tok, err := auth.createToken(uuid.New(), role+"@test.local", role)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	return "Bearer " + tok
}

// ---------------------------------------------------------------------------
// Authentication guard tests
// ---------------------------------------------------------------------------

func TestAssistantChat_Unauthenticated_Returns401(t *testing.T) {
	r, _, _ := setupAssistantRouter(t)
	body := `{"message":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/api/assistant/chat", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAssistantChat_EmptyMessage_Returns400(t *testing.T) {
	r, auth, _ := setupAssistantRouter(t)
	body := `{"message":"   "}`
	req := httptest.NewRequest(http.MethodPost, "/api/assistant/chat", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", assistantToken(t, auth, "patient"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAssistantChat_MissingMessage_Returns400(t *testing.T) {
	r, auth, _ := setupAssistantRouter(t)
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/assistant/chat", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", assistantToken(t, auth, "patient"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Fallback reply tests (no Azure configured)
// ---------------------------------------------------------------------------

func TestAssistantChat_FallbackAppointment(t *testing.T) {
	r, auth, _ := setupAssistantRouter(t)
	body := `{"message":"how do I book an appointment?"}`
	req := httptest.NewRequest(http.MethodPost, "/api/assistant/chat", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", assistantToken(t, auth, "patient"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	reply, _ := res["reply"].(string)
	if reply == "" {
		t.Fatal("expected non-empty reply")
	}
	provider, _ := res["provider"].(string)
	if provider != "rule_fallback" {
		t.Errorf("expected rule_fallback provider, got %q", provider)
	}
	fallbackUsed, _ := res["fallback_used"].(bool)
	if !fallbackUsed {
		t.Error("expected fallback_used=true")
	}
}

func TestAssistantChat_FallbackPrescription(t *testing.T) {
	r, auth, _ := setupAssistantRouter(t)
	body := `{"message":"where are my prescriptions?"}`
	req := httptest.NewRequest(http.MethodPost, "/api/assistant/chat", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", assistantToken(t, auth, "patient"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("parse: %v", err)
	}
	reply, _ := res["reply"].(string)
	if !strings.Contains(strings.ToLower(reply), "prescription") {
		t.Errorf("expected prescription-related reply, got: %q", reply)
	}
}

func TestAssistantChat_FallbackGeneric(t *testing.T) {
	r, auth, _ := setupAssistantRouter(t)
	body := `{"message":"what is the meaning of life?"}`
	req := httptest.NewRequest(http.MethodPost, "/api/assistant/chat", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", assistantToken(t, auth, "doctor"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Multi-turn history tests
// ---------------------------------------------------------------------------

func TestBuildMessages_SystemPromptAlwaysFirst(t *testing.T) {
	history := []ChatMessage{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "hi there"},
	}
	msgs := buildMessages(history, "what can you do?")
	if len(msgs) == 0 || msgs[0].Role != "system" {
		t.Fatal("expected first message to be system")
	}
}

func TestBuildMessages_HistoryPreserved(t *testing.T) {
	history := []ChatMessage{
		{Role: "user", Content: "first message"},
		{Role: "assistant", Content: "first reply"},
	}
	msgs := buildMessages(history, "second message")
	if len(msgs) != 4 { // system + 2 history + 1 new user
		t.Fatalf("expected 4 messages, got %d", len(msgs))
	}
	if msgs[1].Content != "first message" {
		t.Errorf("expected history[0] to be preserved, got %q", msgs[1].Content)
	}
	if msgs[3].Content != "second message" {
		t.Errorf("expected new user message last, got %q", msgs[3].Content)
	}
}

func TestBuildMessages_CapAt20HistoryTurns(t *testing.T) {
	history := make([]ChatMessage, 30)
	for i := range history {
		if i%2 == 0 {
			history[i] = ChatMessage{Role: "user", Content: "msg"}
		} else {
			history[i] = ChatMessage{Role: "assistant", Content: "reply"}
		}
	}
	msgs := buildMessages(history, "new")
	// system(1) + capped 20 history + 1 new = 22
	if len(msgs) != 22 {
		t.Fatalf("expected 22 messages with history cap, got %d", len(msgs))
	}
}

func TestBuildMessages_FiltersInvalidRoles(t *testing.T) {
	history := []ChatMessage{
		{Role: "system", Content: "should be filtered out"},
		{Role: "user", Content: "valid user"},
		{Role: "unknown", Content: "should be filtered"},
		{Role: "assistant", Content: "valid assistant"},
	}
	msgs := buildMessages(history, "new msg")
	for _, m := range msgs[1:] { // skip first system msg
		if m.Role == "unknown" || (m.Role == "system" && m.Content == "should be filtered out") {
			t.Errorf("invalid role %q should have been filtered", m.Role)
		}
	}
}

func TestBuildMessages_EmptyHistory(t *testing.T) {
	msgs := buildMessages(nil, "hello")
	if len(msgs) != 2 { // system + user
		t.Fatalf("expected 2 messages for empty history, got %d", len(msgs))
	}
	if msgs[0].Role != "system" {
		t.Error("expected first message to be system")
	}
	if msgs[1].Role != "user" || msgs[1].Content != "hello" {
		t.Errorf("expected user message 'hello', got role=%q content=%q", msgs[1].Role, msgs[1].Content)
	}
}

// ---------------------------------------------------------------------------
// historyToText tests (Ollama fallback)
// ---------------------------------------------------------------------------

func TestHistoryToText_ProducesReadableFormat(t *testing.T) {
	history := []ChatMessage{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "hi there"},
	}
	out := historyToText(history)
	if !strings.Contains(out, "User: hello") {
		t.Errorf("expected 'User: hello' in output, got: %q", out)
	}
	if !strings.Contains(out, "Assistant: hi there") {
		t.Errorf("expected 'Assistant: hi there' in output, got: %q", out)
	}
}

func TestHistoryToText_EmptyHistory(t *testing.T) {
	out := historyToText(nil)
	if out != "" {
		t.Errorf("expected empty string for nil history, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// smartFallbackReply keyword matching tests
// ---------------------------------------------------------------------------

func TestFallbackAssistantReply_Keywords(t *testing.T) {
	cases := []struct {
		input    string
		contains string
	}{
		{"I need to book an appointment", "appointment"},
		{"Where can I send a message?", "message"},
		{"My document is not loading", "document"},
		{"Find a hospital near me", "hospital"},
		{"I need a prescription renewal", "prescription"},
		{"How do I get notifications?", "notification"},
		{"Random question about something else", "help"},
	}
	for _, tc := range cases {
		reply := smartFallbackReply(tc.input)
		if !strings.Contains(strings.ToLower(reply), tc.contains) {
			t.Errorf("input=%q: expected reply to contain %q, got: %q", tc.input, tc.contains, reply)
		}
	}
}

// ---------------------------------------------------------------------------
// JSON body parsing tests
// ---------------------------------------------------------------------------

func TestAssistantChat_WithHistory_Returns200(t *testing.T) {
	r, auth, _ := setupAssistantRouter(t)
	reqBody := map[string]any{
		"message": "What medications am I on?",
		"history": []map[string]string{
			{"role": "user", "content": "I just registered"},
			{"role": "assistant", "content": "Welcome! How can I help?"},
		},
	}
	raw, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/assistant/chat", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", assistantToken(t, auth, "patient"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if res["reply"] == "" {
		t.Error("expected non-empty reply")
	}
}

func TestAssistantChat_InvalidJSON_Returns400(t *testing.T) {
	r, auth, _ := setupAssistantRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/api/assistant/chat", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", assistantToken(t, auth, "patient"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", w.Code)
	}
}
