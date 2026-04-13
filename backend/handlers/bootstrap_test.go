package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestBootstrap_UnauthorizedWithoutClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/bootstrap", nil)

	h := NewBootstrapHandler()
	h.Get(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestBootstrap_OKWithClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/bootstrap", nil)
	c.Set("claims", &Claims{
		UserID: uuid.New(),
		Email:  "u@b.com",
		Role:   "patient",
	})

	h := NewBootstrapHandler()
	h.Get(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
	}
	var body struct {
		ShellMobileBreakpointPx int `json:"shell_mobile_breakpoint_px"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.ShellMobileBreakpointPx != 768 {
		t.Fatalf("breakpoint: got %d", body.ShellMobileBreakpointPx)
	}
}
