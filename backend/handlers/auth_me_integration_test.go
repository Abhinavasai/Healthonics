package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// TestMe_reloadsFromDatabase verifies GET /api/me returns DB email (F01 contract for account settings).
// Enable with: RUN_AUTH_ME_INTEGRATION=1 DATABASE_URL=postgres://... go test ./handlers -run TestMe_reloadsFromDatabase -count=1
func TestMe_reloadsFromDatabase(t *testing.T) {
	if os.Getenv("RUN_AUTH_ME_INTEGRATION") != "1" {
		t.Skip("set RUN_AUTH_ME_INTEGRATION=1 and DATABASE_URL to run")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL required")
	}

	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	old := db.Pool
	db.Pool = pool
	defer func() { db.Pool = old }()

	if err := db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	suffix := uuid.New().String()[:8]
	email := "me-f01-" + suffix + "@test.local"
	ph, _ := bcrypt.GenerateFromPassword([]byte("testpass123"), bcrypt.DefaultCost)
	var uid uuid.UUID
	err = pool.QueryRow(ctx, `INSERT INTO users (email, password_hash, role) VALUES ($1, $2, 'patient') RETURNING id`,
		email, string(ph)).Scan(&uid)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, uid)
	})

	secret := []byte("integration-secret-me-f01")
	h := NewAuthHandler(string(secret))
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		UserID: uid,
		Email:  "stale-wrong@example.com",
		Role:   "patient",
	})
	signed, err := tok.SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/me", nil)
	c.Request.Header.Set("Authorization", "Bearer "+signed)
	c.Set("claims", &Claims{UserID: uid, Email: "stale-wrong@example.com", Role: "patient"})

	// Simulate RequireAuth having set claims; Me only reads claims + DB.
	h.Me(c)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /me: %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), email) {
		t.Fatalf("response should include DB email %q, body=%s", email, w.Body.String())
	}
}
