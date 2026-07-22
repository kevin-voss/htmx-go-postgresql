package app

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/config"
	"github.com/kevin-voss/htmx-go-postgresql/internal/mail"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/render"
	"github.com/kevin-voss/htmx-go-postgresql/web"
)

// Application holds process dependencies for the HTTP server.
type Application struct {
	Config config.Config
	Logger *slog.Logger
	DB     *pgxpool.Pool
	Render *render.Renderer
	Auth   *auth.Handler
}

// New constructs an Application with the given config, logger, and database pool.
func New(cfg config.Config, logger *slog.Logger, db *pgxpool.Pool) *Application {
	renderer, err := render.New(web.Templates)
	if err != nil {
		// Templates are embedded at build time; fail fast if parsing drifts.
		panic("render templates: " + err.Error())
	}

	mailer := mail.Sender(mail.NopMailer{})
	if cfg.SMTPHost != "" && cfg.SMTPPort != "" {
		smtpMailer, err := mail.NewSMTP(cfg.SMTPHost, cfg.SMTPPort)
		if err != nil {
			panic("mail smtp: " + err.Error())
		}
		mailer = smtpMailer
	}

	repo := auth.NewRepository(db)
	authHandler := auth.NewHandler(
		auth.NewService(repo, repo, repo),
		mailer,
		renderer,
		logger,
		cfg.CookieSecure,
	)

	return &Application{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Render: renderer,
		Auth:   authHandler,
	}
}
