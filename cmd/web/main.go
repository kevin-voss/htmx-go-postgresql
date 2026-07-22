package main

import (
	"log/slog"
	"os"

	"github.com/kevin-voss/htmx-go-postgresql/internal/config"
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
}
