package config

import (
	"fmt"
	"os"
	"strings"
)

// defaultLocalPostgresURL matches docker-compose postgres service (app/app on localhost:5432).
const defaultLocalPostgresURL = "postgres://app:app@127.0.0.1:5432/app?sslmode=disable"

type Config struct {
	HTTPAddr         string
	DBURL            string
	JWTSecret        string
	GroqAPIKey       string
	GroqModel        string
	StorageProvider  string
	StorageLocalDir  string
	GCSBucket           string
	GCSReportsPrefix    string
	PubSubProjectID     string
	PubSubTopicID       string
	PubSubSubscriptionID string
	CORSAllowedOrigins  []string
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	config := Config{
		HTTPAddr:  envOrDefault("HTTP_ADDR", ":8080"),
		JWTSecret: os.Getenv("JWT_SECRET"),
	}

	if config.HTTPAddr == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR must not be empty")
	}

	config.DBURL = os.Getenv("DATABASE_URL")
	if config.DBURL == "" {
		config.DBURL = defaultLocalPostgresURL
	}

	config.GroqAPIKey = os.Getenv("GROQ_API_KEY")
	config.GroqModel = envOrDefault("GROQ_MODEL", "llama-3.1-8b-instant")
	config.StorageProvider = envOrDefault("STORAGE_PROVIDER", "local")
	config.StorageLocalDir = envOrDefault("STORAGE_LOCAL_DIR", "./uploads")
	config.GCSBucket = os.Getenv("GCS_BUCKET")
	config.GCSReportsPrefix = envOrDefault("GCS_REPORTS_PREFIX", "reports")
	config.PubSubProjectID = envOrDefault("GCP_PROJECT_ID", "local-project")
	config.PubSubTopicID = envOrDefault("PUBSUB_TOPIC_ID", "reports")
	config.PubSubSubscriptionID = envOrDefault("PUBSUB_SUBSCRIPTION_ID", "reports-weekly-generate")
	config.CORSAllowedOrigins = parseCSV(os.Getenv("CORS_ALLOWED_ORIGINS"))

	return config, nil
}

func parseCSV(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
