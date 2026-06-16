package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
	"github.com/healthonyx/backend/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	JWTSecret []byte
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthHandler(jwtSecret string) *AuthHandler {
	return &AuthHandler{JWTSecret: []byte(jwtSecret)}
}

func validRole(role string) bool {
	return role == "patient" || role == "doctor" || role == "admin"
}

func validPassword(pw string) bool {
	return len([]rune(pw)) >= 6
}

func validEmailFormat(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil && strings.Contains(email, ".")
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}
	if !validEmailFormat(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}
	if !validPassword(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
		return
	}
	if !validRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role must be patient, doctor, or admin"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	var id uuid.UUID
	err = db.Pool.QueryRow(c.Request.Context(),
		`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3) RETURNING id`,
		req.Email, string(hash), req.Role,
	).Scan(&id)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	token, err := h.createToken(id, req.Email, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	refreshToken, _ := h.issueRefreshToken(c, id) // best-effort; login still succeeds if this fails

	c.JSON(http.StatusCreated, gin.H{
		"token":         token,
		"refresh_token": refreshToken,
		"user":          models.UserProfile{ID: id, Email: req.Email, Role: req.Role},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	var id uuid.UUID
	var passwordHash, role string
	var isActive bool
	err := db.Pool.QueryRow(c.Request.Context(),
		`SELECT id, password_hash, role, is_active FROM users WHERE email = $1`, req.Email,
	).Scan(&id, &passwordHash, &role, &isActive)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if !isActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is deactivated"})
		return
	}

	token, err := h.createToken(id, req.Email, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	refreshToken, _ := h.issueRefreshToken(c, id)

	c.JSON(http.StatusOK, gin.H{
		"token":         token,
		"refresh_token": refreshToken,
		"user":          models.UserProfile{ID: id, Email: req.Email, Role: role},
	})
}

// generateRefreshToken creates a cryptographically random 256-bit token.
func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// issueRefreshToken stores a new refresh token and returns it.
// Refresh tokens are long-lived (30 days) and stored server-side for revocation.
func (h *AuthHandler) issueRefreshToken(c *gin.Context, userID uuid.UUID) (string, error) {
	token, err := generateRefreshToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().UTC().Add(30 * 24 * time.Hour)
	_, err = db.Pool.Exec(c.Request.Context(), `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, userID, token, expiresAt)
	if err != nil {
		return "", err
	}
	return token, nil
}

// Refresh  POST /api/auth/refresh
// Validates a refresh token and issues a new short-lived access token.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.RefreshToken) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	var userID uuid.UUID
	var email, role string
	var isActive bool
	err := db.Pool.QueryRow(c.Request.Context(), `
		SELECT u.id, u.email, u.role, u.is_active
		FROM refresh_tokens rt
		JOIN users u ON u.id = rt.user_id
		WHERE rt.token = $1
		  AND rt.revoked = FALSE
		  AND rt.expires_at > NOW()
	`, strings.TrimSpace(body.RefreshToken)).Scan(&userID, &email, &role, &isActive)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}
	if !isActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is deactivated"})
		return
	}

	// Rotate: revoke old token, issue new refresh + access tokens.
	_, _ = db.Pool.Exec(c.Request.Context(), `UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1`, body.RefreshToken)

	newRefresh, err := h.issueRefreshToken(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	accessToken, err := h.createToken(userID, email, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         accessToken,
		"refresh_token": newRefresh,
		"user":          models.UserProfile{ID: userID, Email: email, Role: role},
	})
}

// Logout  POST /api/auth/logout — revokes the caller's refresh token.
func (h *AuthHandler) Logout(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&body)
	if t := strings.TrimSpace(body.RefreshToken); t != "" {
		_, _ = db.Pool.Exec(c.Request.Context(), `UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1`, t)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AuthHandler) Me(c *gin.Context) {
	claimsVal, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims := claimsVal.(*Claims)
	c.JSON(http.StatusOK, models.UserProfile{ID: claims.UserID, Email: claims.Email, Role: claims.Role})
}

func (h *AuthHandler) createToken(id uuid.UUID, email, role string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: id,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-1 * time.Minute)),
			ExpiresAt: jwt.NewNumericDate(now.Add(12 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.JWTSecret)
}

func (h *AuthHandler) parseClaims(rawToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(rawToken, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return h.JWTSecret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}

// ParseJWTClaims validates a raw JWT string (used by WebSocket token auth).
func (h *AuthHandler) ParseJWTClaims(rawToken string) (*Claims, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil, errors.New("token required")
	}
	return h.parseClaims(rawToken)
}

func (h *AuthHandler) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header"})
			c.Abort()
			return
		}
		claims, err := h.parseClaims(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}

func (h *AuthHandler) RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool)
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		claimsVal, ok := c.Get("claims")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		claims := claimsVal.(*Claims)
		if !allowed[claims.Role] {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}
