package config

import (
	"fmt"
	"os"
)

// defaultLocalPostgresURL matches docker-compose postgres service (app/app on localhost:5432).
const defaultLocalPostgresURL = "postgres://app:app@127.0.0.1:5432/app?sslmode=disable"

type Config struct {
	HTTPAddr  string
	DBURL     string
	JWTSecret string
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

	return config, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
