package app

import (
	"log/slog"

	"github.com/kevin-voss/htmx-go-postgresql/internal/config"
)

// Application holds process dependencies for the HTTP server.
type Application struct {
	Config config.Config
	Logger *slog.Logger
}

// New constructs an Application with the given config and logger.
func New(cfg config.Config, logger *slog.Logger) *Application {
	return &Application{
		Config: cfg,
		Logger: logger,
	}
}
