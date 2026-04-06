package config

import (
	"fmt"
	"os"
)

// URL por defecto cuando corres el API en tu PC y Postgres está en Docker (puerto 5432).
const postgresLocalPorDefecto = "postgres://app:app@127.0.0.1:5432/app?sslmode=disable"

type Config struct {
	HTTPAddr string
	DBURL    string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr: envConDefecto("HTTP_ADDR", ":8080"),
	}

	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR no puede estar vacío")
	}

	cfg.DBURL = os.Getenv("DATABASE_URL")
	if cfg.DBURL == "" {
		cfg.DBURL = postgresLocalPorDefecto
	}

	return cfg, nil
}

func envConDefecto(nombre, defecto string) string {
	if v := os.Getenv(nombre); v != "" {
		return v
	}
	return defecto
}
