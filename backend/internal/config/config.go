package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config aggregates runtime configuration knobs.
type Config struct {
	HTTPPort      string
	DatabaseURL   string
	AllowedOrigin string
	QuestionCount int
}

// Load reads environment variables and produces a Config.
func Load() (Config, error) {
	cfg := Config{
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		AllowedOrigin: os.Getenv("ALLOWED_ORIGIN"),
		QuestionCount: getEnvAsInt("QUESTION_COUNT", 92),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL must be set")
	}

	if cfg.QuestionCount <= 0 {
		return Config{}, fmt.Errorf("QUESTION_COUNT must be positive")
	}
	if cfg.QuestionCount%4 != 0 {
		return Config{}, fmt.Errorf("QUESTION_COUNT must be divisible by four")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return fallback
}
