package app

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kevin-voss/htmx-go-postgresql/internal/config"
)

// Application holds process dependencies for the HTTP server.
type Application struct {
	Config config.Config
	Logger *slog.Logger
	DB     *pgxpool.Pool
}

// New constructs an Application with the given config, logger, and database pool.
func New(cfg config.Config, logger *slog.Logger, db *pgxpool.Pool) *Application {
	return &Application{
		Config: cfg,
		Logger: logger,
		DB:     db,
	}
}
