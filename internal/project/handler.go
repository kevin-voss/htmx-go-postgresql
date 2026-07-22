package project

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

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
	list := middleware.Chain(
		http.HandlerFunc(h.list),
		member.RequireMembership(h.members, h.logger),
	)
	mux.Handle("GET /w/{workspaceSlug}/projects", auth.RequireAuthentication(list))

	create := middleware.Chain(
		http.HandlerFunc(h.create),
		member.RequireMembership(h.members, h.logger),
		member.RequireCanMutate(),
	)
	mux.Handle("POST /w/{workspaceSlug}/projects", auth.RequireAuthentication(create))

	show := middleware.Chain(
		http.HandlerFunc(h.show),
		member.RequireMembership(h.members, h.logger),
	)
	mux.Handle("GET /w/{workspaceSlug}/projects/{projectSlug}", auth.RequireAuthentication(show))
}

type listPageData struct {
	CSRFToken     string
	WorkspaceID   string
	WorkspaceName string
	WorkspaceSlug string
	Projects      []Project
	User          auth.User
	Role          string
	CanCreate     bool
	Form          createFormData
	Errors        CreateErrors
}

type createFormData struct {
	Name string
	Slug string
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

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	ws, ok := workspaceFromAccessContext(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	role, _ := middleware.WorkspaceRoleFromContext(r.Context())

	projects, err := h.service.ListByWorkspace(r.Context(), ws.ID)
	if err != nil {
		h.logger.Error("list projects failed", "err", err, "workspace_id", ws.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.renderList(w, http.StatusOK, listPageData{
		CSRFToken:     middleware.CSRFToken(r.Context()),
		WorkspaceID:   ws.ID,
		WorkspaceName: ws.Name,
		WorkspaceSlug: ws.Slug,
		Projects:      projects,
		User:          user,
		Role:          role,
		CanCreate:     member.Role(role).CanMutate(),
	})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	ws, ok := workspaceFromAccessContext(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	role, _ := middleware.WorkspaceRoleFromContext(r.Context())

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	p, fieldErrs, err := h.service.Create(r.Context(), CreateInput{
		WorkspaceID: ws.ID,
		Name:        r.FormValue("name"),
		Slug:        r.FormValue("slug"),
		CreatedBy:   user.ID,
	})
	if err != nil {
		h.logger.Error("create project failed", "err", err, "workspace_id", ws.ID, "user_id", user.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if fieldErrs.Any() {
		projects, listErr := h.service.ListByWorkspace(r.Context(), ws.ID)
		if listErr != nil {
			h.logger.Error("list projects failed", "err", listErr, "workspace_id", ws.ID)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		h.renderList(w, http.StatusUnprocessableEntity, listPageData{
			CSRFToken:     middleware.CSRFToken(r.Context()),
			WorkspaceID:   ws.ID,
			WorkspaceName: ws.Name,
			WorkspaceSlug: ws.Slug,
			Projects:      projects,
			User:          user,
			Role:          role,
			CanCreate:     true,
			Form: createFormData{
				Name: strings.TrimSpace(r.FormValue("name")),
				Slug: strings.ToLower(strings.TrimSpace(r.FormValue("slug"))),
			},
			Errors: fieldErrs,
		})
		return
	}

	h.logger.Info(
		"project created",
		"project_id", p.ID,
		"slug", p.Slug,
		"workspace_id", ws.ID,
		"user_id", user.ID,
	)
	http.Redirect(w, r, "/w/"+ws.Slug+"/projects/"+p.Slug, http.StatusSeeOther)
}

func (h *Handler) show(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	ws, ok := workspaceFromAccessContext(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	role, _ := middleware.WorkspaceRoleFromContext(r.Context())

	p, err := h.service.GetByWorkspaceAndSlug(r.Context(), ws.ID, r.PathValue("projectSlug"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("get project failed", "err", err, "workspace_id", ws.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := h.render.Render(w, http.StatusOK, "project_show", showPageData{
		CSRFToken:     middleware.CSRFToken(r.Context()),
		WorkspaceID:   ws.ID,
		WorkspaceName: ws.Name,
		WorkspaceSlug: ws.Slug,
		Project:       p,
		User:          user,
		Role:          role,
	}); err != nil {
		h.logger.Error("render project show failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

type workspaceAccess struct {
	ID   string
	Name string
	Slug string
}

func workspaceFromAccessContext(r *http.Request) (workspaceAccess, bool) {
	id, okID := middleware.WorkspaceIDFromContext(r.Context())
	name, okName := middleware.WorkspaceNameFromContext(r.Context())
	slug, okSlug := middleware.WorkspaceSlugFromContext(r.Context())
	if !okID || !okName || !okSlug {
		return workspaceAccess{}, false
	}
	return workspaceAccess{ID: id, Name: name, Slug: slug}, true
}

func (h *Handler) renderList(w http.ResponseWriter, status int, data listPageData) {
	if err := h.render.Render(w, status, "project_list", data); err != nil {
		h.logger.Error("render project list failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
