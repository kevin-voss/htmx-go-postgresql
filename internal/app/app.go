package app

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/config"
	"github.com/kevin-voss/htmx-go-postgresql/internal/issue"
	"github.com/kevin-voss/htmx-go-postgresql/internal/mail"
	"github.com/kevin-voss/htmx-go-postgresql/internal/member"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/render"
	"github.com/kevin-voss/htmx-go-postgresql/internal/project"
	"github.com/kevin-voss/htmx-go-postgresql/internal/workspace"
	"github.com/kevin-voss/htmx-go-postgresql/web"
)

// Application holds process dependencies for the HTTP server.
type Application struct {
	Config     config.Config
	Logger     *slog.Logger
	DB         *pgxpool.Pool
	Render     *render.Renderer
	Auth       *auth.Handler
	Workspace  *workspace.Handler
	Project    *project.Handler
	Issue      *issue.Handler
	Members    *member.Service
	MemberHTTP *member.Handler
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
		auth.NewService(repo, repo, repo, repo),
		mailer,
		renderer,
		logger,
		cfg.CookieSecure,
	)

	memberRepo := member.NewRepository(db)
	memberService := member.NewService(memberRepo).WithUserLookup(member.AuthUserLookup{Users: repo})
	memberHandler := member.NewHandler(memberService, mailer, renderer, logger)

	workspaceRepo := workspace.NewRepository(db)
	workspaceHandler := workspace.NewHandler(
		workspace.NewService(workspaceRepo),
		memberService,
		renderer,
		logger,
	)

	projectRepo := project.NewRepository(db)
	projectService := project.NewService(projectRepo)
	projectHandler := project.NewHandler(
		projectService,
		memberService,
		renderer,
		logger,
	)

	issueRepo := issue.NewRepository(db)
	issueService := issue.NewService(issueRepo).WithMembershipChecker(memberService)
	issueHandler := issue.NewHandler(
		issueService,
		projectService,
		memberService,
		renderer,
		logger,
	)

	return &Application{
		Config:     cfg,
		Logger:     logger,
		DB:         db,
		Render:     renderer,
		Auth:       authHandler,
		Workspace:  workspaceHandler,
		Project:    projectHandler,
		Issue:      issueHandler,
		Members:    memberService,
		MemberHTTP: memberHandler,
	}
}
