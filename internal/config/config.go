package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	HTTPAddr    string
	LogLevel    string

	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		HTTPAddr:    getEnvDefault("HTTP_ADDR", ":8080"),
		LogLevel:    getEnvDefault("LOG_LEVEL", "info"),

		DBMaxOpenConns:    getEnvIntDefault("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getEnvIntDefault("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetime: getEnvDurationDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvIntDefault(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getEnvDurationDefault(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
