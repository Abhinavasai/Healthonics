package config

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

type Config struct {
	DatabaseURL       string
	JWTSecret         string
	Port              string
	CORSOrigins       string // Comma-separated, e.g. "http://localhost:4200,http://localhost:3000"
	NominatimBaseURL  string // OpenStreetMap Nominatim (or self-hosted). Empty = public default.
	GeocodeUserAgent  string // Required-style identification for Nominatim; override in production.
	UploadDir         string // Patient file uploads (filesystem); default data/uploads
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
	return &Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		Port:             port,
		CORSOrigins:      os.Getenv("CORS_ORIGINS"),
		NominatimBaseURL: os.Getenv("NOMINATIM_BASE_URL"),
		GeocodeUserAgent: os.Getenv("GEOCODE_USER_AGENT"),
		UploadDir:        uploadDir,
	}
}
