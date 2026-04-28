package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

type Config struct {
	DatabaseURL      string
	JWTSecret        string
	Port             string
	CORSOrigins      string // Comma-separated, e.g. "http://localhost:4200,http://localhost:3000"
	NominatimBaseURL string // OpenStreetMap Nominatim (or self-hosted). Empty = public default.
	GeocodeUserAgent string // Required-style identification for Nominatim; override in production.
	UploadDir        string // Patient file uploads (filesystem); default data/uploads
	AIEnabled        bool   // Global AI kill switch. false => always use fallback behavior.
	OllamaHost       string // Ollama base URL, e.g. http://127.0.0.1:11434
	OllamaModel      string // Default Ollama model name.
	OllamaTimeoutMS  int    // Timeout for Ollama HTTP calls in milliseconds.
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "data/uploads"
	}
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://127.0.0.1:11434"
	}
	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "llama3.1:8b"
	}
	return &Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		Port:             port,
		CORSOrigins:      os.Getenv("CORS_ORIGINS"),
		NominatimBaseURL: os.Getenv("NOMINATIM_BASE_URL"),
		GeocodeUserAgent: os.Getenv("GEOCODE_USER_AGENT"),
		UploadDir:        uploadDir,
		AIEnabled:        strings.EqualFold(strings.TrimSpace(os.Getenv("AI_ENABLED")), "true"),
		OllamaHost:       ollamaHost,
		OllamaModel:      ollamaModel,
		OllamaTimeoutMS:  parsePositiveIntWithDefault(os.Getenv("OLLAMA_TIMEOUT_MS"), 7000),
	}
}

func parsePositiveIntWithDefault(raw string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}
