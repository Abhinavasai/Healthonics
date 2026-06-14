package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

type Config struct {
	DatabaseURL           string
	JWTSecret             string
	Port                  string
	CORSOrigins           string // Comma-separated, e.g. "http://localhost:4200,http://localhost:3000"
	NominatimBaseURL      string // OpenStreetMap Nominatim (or self-hosted). Empty = public default.
	GeocodeUserAgent      string // Required-style identification for Nominatim; override in production.
	UploadDir             string // Patient file uploads (filesystem); default data/uploads
	AIEnabled             bool   // Global AI kill switch. false => always use fallback behavior.
	OllamaHost            string // Ollama base URL, e.g. http://127.0.0.1:11434
	OllamaModel           string // Default Ollama model name.
	OllamaTimeoutMS       int    // Timeout for Ollama HTTP calls in milliseconds.
	NotifyIntervalMS      int    // Worker polling interval for notification queue.
	NotifyMaxRetries      int    // Maximum automatic retries before dead-letter.
	SendGridAPIKey        string // Optional SendGrid API key.
	SendGridFrom          string // Optional sender email for SendGrid.
	SendGridWebhookSecret string // Optional webhook HMAC secret for SendGrid callbacks.
	TwilioAccountSID      string // Optional Twilio account SID.
	TwilioAuthToken       string // Optional Twilio auth token.
	TwilioFromNumber      string // Optional Twilio sender number.
	TwilioWebhookSecret   string // Optional webhook HMAC secret for Twilio callbacks.
	GoogleMapsAPIKey      string // Optional Google Maps/Places API key for geocode + nearby hospitals.
	AzureOpenAIEndpoint   string // Optional Azure OpenAI endpoint.
	AzureOpenAIAPIKey     string // Optional Azure OpenAI API key.
	AzureOpenAIAPIVersion string // Optional Azure OpenAI API version.
	AzureOpenAIModel      string // Optional Azure OpenAI model/deployment name.
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
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		JWTSecret:             os.Getenv("JWT_SECRET"),
		Port:                  port,
		CORSOrigins:           os.Getenv("CORS_ORIGINS"),
		NominatimBaseURL:      os.Getenv("NOMINATIM_BASE_URL"),
		GeocodeUserAgent:      os.Getenv("GEOCODE_USER_AGENT"),
		UploadDir:             uploadDir,
		AIEnabled:             strings.EqualFold(strings.TrimSpace(os.Getenv("AI_ENABLED")), "true"),
		OllamaHost:            ollamaHost,
		OllamaModel:           ollamaModel,
		OllamaTimeoutMS:       parsePositiveIntWithDefault(os.Getenv("OLLAMA_TIMEOUT_MS"), 7000),
		NotifyIntervalMS:      parsePositiveIntWithDefault(os.Getenv("NOTIFY_WORKER_INTERVAL_MS"), 5000),
		NotifyMaxRetries:      parseIntInRangeWithDefault(os.Getenv("NOTIFY_MAX_RETRIES"), 1, 20, 3),
		SendGridAPIKey:        strings.TrimSpace(os.Getenv("SENDGRID_API_KEY")),
		SendGridFrom:          strings.TrimSpace(os.Getenv("SENDGRID_FROM_EMAIL")),
		SendGridWebhookSecret: strings.TrimSpace(os.Getenv("SENDGRID_WEBHOOK_SECRET")),
		TwilioAccountSID:      strings.TrimSpace(os.Getenv("TWILIO_ACCOUNT_SID")),
		TwilioAuthToken:       strings.TrimSpace(os.Getenv("TWILIO_AUTH_TOKEN")),
		TwilioFromNumber:      strings.TrimSpace(os.Getenv("TWILIO_FROM_NUMBER")),
		TwilioWebhookSecret:   strings.TrimSpace(os.Getenv("TWILIO_WEBHOOK_SECRET")),
		GoogleMapsAPIKey:      strings.TrimSpace(os.Getenv("GOOGLE_MAPS_API_KEY")),
		AzureOpenAIEndpoint:   strings.TrimSpace(os.Getenv("AZURE_OPENAI_ENDPOINT")),
		AzureOpenAIAPIKey:     strings.TrimSpace(os.Getenv("AZURE_OPENAI_API_KEY")),
		AzureOpenAIAPIVersion: strings.TrimSpace(os.Getenv("AZURE_OPENAI_API_VERSION")),
		AzureOpenAIModel:      strings.TrimSpace(os.Getenv("AZURE_OPENAI_MODEL")),
	}
}

// Validate returns an error if any required configuration is missing or invalid.
// Call this at startup so the service fails fast rather than dying mid-request.
func (c *Config) Validate() error {
	var missing []string
	if c.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if len(c.JWTSecret) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters")
	}
	if len(missing) > 0 {
		return errors.New("required environment variables not set: " + strings.Join(missing, ", "))
	}
	return nil
}

func parsePositiveIntWithDefault(raw string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func parseIntInRangeWithDefault(raw string, min int, max int, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v < min || v > max {
		return fallback
	}
	return v
}
