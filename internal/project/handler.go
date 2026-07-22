package project

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/member"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/render"
)

// Handler serves project HTTP endpoints.
type Handler struct {
	service *Service
	members *member.Service
	render  *render.Renderer
	logger  *slog.Logger
}

// NewHandler constructs a project HTTP handler.
func NewHandler(service *Service, members *member.Service, renderer *render.Renderer, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		members: members,
		render:  renderer,
		logger:  logger,
	}
}

// Mount registers project routes on mux (all require authentication + membership).
func (h *Handler) Mount(mux *http.ServeMux) {
	show := middleware.Chain(
		http.HandlerFunc(h.show),
		member.RequireMembership(h.members, h.logger),
	)
	mux.Handle("GET /w/{workspaceSlug}/projects/{projectSlug}", auth.RequireAuthentication(show))
}

type showPageData struct {
	CSRFToken     string
	WorkspaceID   string
	WorkspaceName string
	WorkspaceSlug string
	Project       Project
	User          auth.User
	Role          string
}

func (h *Handler) show(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	workspaceID, ok := middleware.WorkspaceIDFromContext(r.Context())
	if !ok {
		http.NotFound(w, r)
		return
	}
	workspaceName, _ := middleware.WorkspaceNameFromContext(r.Context())
	workspaceSlug, _ := middleware.WorkspaceSlugFromContext(r.Context())
	role, _ := middleware.WorkspaceRoleFromContext(r.Context())

	p, err := h.service.GetByWorkspaceAndSlug(r.Context(), workspaceID, r.PathValue("projectSlug"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("get project failed", "err", err, "workspace_id", workspaceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := h.render.Render(w, http.StatusOK, "project_show", showPageData{
		CSRFToken:     middleware.CSRFToken(r.Context()),
		WorkspaceID:   workspaceID,
		WorkspaceName: workspaceName,
		WorkspaceSlug: workspaceSlug,
		Project:       p,
		User:          user,
		Role:          role,
	}); err != nil {
		h.logger.Error("render project show failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
