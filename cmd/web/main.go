package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/kevin-voss/htmx-go-postgresql/internal/app"
	"github.com/kevin-voss/htmx-go-postgresql/internal/config"
	"github.com/kevin-voss/htmx-go-postgresql/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "err", err)
		os.Exit(1)
	}

	logger := config.NewLogger(cfg)
	slog.SetDefault(logger)

	slog.Info("forgeboard starting",
		"env", cfg.Env,
		"address", cfg.Address,
	)

	db, err := database.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("database open failed", "err", err)
		os.Exit(1)
	}

	application := app.New(cfg, logger, db)
	if err := application.Run(context.Background()); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
