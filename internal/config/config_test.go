package config

import (
	"testing"
)

func TestLoadDefaultsAndValidation(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("APP_ADDRESS", "")
	t.Setenv("DATABASE_URL", "postgres://forgeboard:forgeboard@localhost:5432/forgeboard?sslmode=disable")
	t.Setenv("SMTP_HOST", "localhost")
	t.Setenv("SMTP_PORT", "1025")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Address != ":8080" {
		t.Fatalf("Address = %q, want :8080", cfg.Address)
	}
	if cfg.Env != "development" {
		t.Fatalf("Env = %q, want development", cfg.Env)
	}
	if cfg.CookieSecure {
		t.Fatal("CookieSecure should be false in development")
	}
	if cfg.SMTPHost != "localhost" || cfg.SMTPPort != "1025" {
		t.Fatalf("SMTP = %s:%s, want localhost:1025", cfg.SMTPHost, cfg.SMTPPort)
	}
}

func TestLoadRequiresDatabaseURLOutsideTest(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want DATABASE_URL required")
	}
}

func TestLoadAllowsMissingDatabaseURLInTestEnv(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("DATABASE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.DatabaseURL != "" {
		t.Fatalf("DatabaseURL = %q, want empty", cfg.DatabaseURL)
	}
}

func TestLoadCookieSecureInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://example")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if !cfg.CookieSecure {
		t.Fatal("CookieSecure should be true in production")
	}
}
