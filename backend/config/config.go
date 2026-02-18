package config

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

type Config struct {
	DatabaseURL  string
	JWTSecret    string
	Port         string
	CORSOrigins  string // Comma-separated, e.g. "http://localhost:4200,http://localhost:3000"
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Port:        port,
		CORSOrigins: os.Getenv("CORS_ORIGINS"),
	}
}
