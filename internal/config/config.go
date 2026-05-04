package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	APIPort     string
}

// Load читает переменные окружения из .env файла.
// Если .env не найден — берёт переменные из окружения (для Docker/K8s).
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
		APIPort:     os.Getenv("API_PORT"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	if cfg.RedisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is not set")
	}
	if cfg.APIPort == "" {
		cfg.APIPort = "8080" // дефолт
	}

	return cfg, nil
}
