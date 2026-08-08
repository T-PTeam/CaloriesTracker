package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr            string
	DatabaseURL         string
	OpenAIAPIKey        string
	OpenAIModel         string
	TelegramBotToken    string
	TelegramWebhookPath string
	APISecret           string
	JWTSecret           string
	CORSAllowedOrigins  []string
	WebAppURL           string
	ShutdownTimeout     time.Duration
	RequestTimeout      time.Duration
}

func Load() (Config, error) {
	apiSecret := os.Getenv("API_SECRET")
	cfg := Config{
		HTTPAddr:            getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		OpenAIAPIKey:        os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:         getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		TelegramBotToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramWebhookPath: getEnv("TELEGRAM_WEBHOOK_PATH", "/telegram/webhook"),
		APISecret:           apiSecret,
		JWTSecret:           getEnv("JWT_SECRET", apiSecret),
		CORSAllowedOrigins:  parseCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")),
		WebAppURL:           getEnv("WEB_APP_URL", "https://calories.fittracker.store"),
		ShutdownTimeout:     getDuration("SHUTDOWN_TIMEOUT", 15*time.Second),
		RequestTimeout:      getDuration("REQUEST_TIMEOUT", 30*time.Second),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.OpenAIAPIKey == "" {
		return Config{}, fmt.Errorf("OPENAI_API_KEY is required")
	}
	if cfg.TelegramBotToken == "" {
		return Config{}, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.APISecret == "" {
		return Config{}, fmt.Errorf("API_SECRET is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func parseCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}
