package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Config holds process settings loaded from the environment.
// It is intended to be injected into the application struct in later steps.
type Config struct {
	Env          string // APP_ENV: development | production | test
	Address      string // APP_ADDRESS
	DatabaseURL  string // DATABASE_URL
	SMTPHost     string // SMTP_HOST
	SMTPPort     string // SMTP_PORT
	CookieSecure bool   // Secure cookies (true in production)
}

// Load reads configuration from environment variables.
// DATABASE_URL is required unless APP_ENV is "test".
// APP_ADDRESS defaults to ":8080" when unset.
func Load() (Config, error) {
	env := strings.TrimSpace(os.Getenv("APP_ENV"))
	if env == "" {
		env = "development"
	}

	addr := strings.TrimSpace(os.Getenv("APP_ADDRESS"))
	if addr == "" {
		addr = ":8080"
	}

	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dbURL == "" && env != "test" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	smtpHost := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	smtpPort := strings.TrimSpace(os.Getenv("SMTP_PORT"))

	cfg := Config{
		Env:          env,
		Address:      addr,
		DatabaseURL:  dbURL,
		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
		CookieSecure: env == "production",
	}
	return cfg, nil
}

// NewLogger returns a structured slog logger sized for APP_ENV.
// Development uses text + debug; production uses JSON + info.
func NewLogger(cfg Config) *slog.Logger {
	var level slog.Level
	switch cfg.Env {
	case "production":
		level = slog.LevelInfo
	case "test":
		level = slog.LevelError
	default:
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if cfg.Env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}
