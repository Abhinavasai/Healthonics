package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func setupProtectedRoleRouter(t *testing.T) (*gin.Engine, *AuthHandler) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	auth := NewAuthHandler("test-secret")
	r := gin.New()
	r.GET("/api/protected/patient", auth.RequireAuth(), auth.RequireRole("patient"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/api/protected/doctor", auth.RequireAuth(), auth.RequireRole("doctor"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/api/protected/admin", auth.RequireAuth(), auth.RequireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r, auth
}

func bearerToken(t *testing.T, auth *AuthHandler, role string) string {
	t.Helper()
	token, err := auth.createToken(uuid.New(), role+"@test.local", role)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	return "Bearer " + token
}

func TestRequireAuthRole_NoAuthHeader_Returns401(t *testing.T) {
	r, _ := setupProtectedRoleRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/protected/patient", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuthRole_InvalidToken_Returns401(t *testing.T) {
	r, _ := setupProtectedRoleRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/protected/patient", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuthRole_WrongRole_Returns403(t *testing.T) {
	r, auth := setupProtectedRoleRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/protected/admin", nil)
	req.Header.Set("Authorization", bearerToken(t, auth, "doctor"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestRequireAuthRole_AllowedRole_Returns200(t *testing.T) {
	r, auth := setupProtectedRoleRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/protected/doctor", nil)
	req.Header.Set("Authorization", bearerToken(t, auth, "doctor"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequireAuthRole_MultiRole_AllowsEither(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := NewAuthHandler("test-secret")
	r := gin.New()
	r.GET("/api/protected/doctor-admin", auth.RequireAuth(), auth.RequireRole("doctor", "admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	doctorReq := httptest.NewRequest(http.MethodGet, "/api/protected/doctor-admin", nil)
	doctorReq.Header.Set("Authorization", bearerToken(t, auth, "doctor"))
	doctorW := httptest.NewRecorder()
	r.ServeHTTP(doctorW, doctorReq)
	if doctorW.Code != http.StatusOK {
		t.Fatalf("expected doctor to be allowed (200), got %d", doctorW.Code)
	}

	patientReq := httptest.NewRequest(http.MethodGet, "/api/protected/doctor-admin", nil)
	patientReq.Header.Set("Authorization", bearerToken(t, auth, "patient"))
	patientW := httptest.NewRecorder()
	r.ServeHTTP(patientW, patientReq)
	if patientW.Code != http.StatusForbidden {
		t.Fatalf("expected patient forbidden (403), got %d", patientW.Code)
	}
}
