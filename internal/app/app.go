package app

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kevin-voss/htmx-go-postgresql/internal/config"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/render"
	"github.com/kevin-voss/htmx-go-postgresql/web"
)

// Application holds process dependencies for the HTTP server.
type Application struct {
	Config config.Config
	Logger *slog.Logger
	DB     *pgxpool.Pool
	Render *render.Renderer
}

// New constructs an Application with the given config, logger, and database pool.
func New(cfg config.Config, logger *slog.Logger, db *pgxpool.Pool) *Application {
	renderer, err := render.New(web.Templates)
	if err != nil {
		// Templates are embedded at build time; fail fast if parsing drifts.
		panic("render templates: " + err.Error())
	}

	return &Application{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Render: renderer,
	}
}
