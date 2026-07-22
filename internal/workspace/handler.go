package workspace

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/member"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/render"
)

// Handler serves workspace HTTP endpoints.
type Handler struct {
	service *Service
	members *member.Service
	render  *render.Renderer
	logger  *slog.Logger
}

// NewHandler constructs a workspace HTTP handler.
func NewHandler(service *Service, members *member.Service, renderer *render.Renderer, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		members: members,
		render:  renderer,
		logger:  logger,
	}
}

// Mount registers workspace routes on mux (all require authentication).
func (h *Handler) Mount(mux *http.ServeMux) {
	mux.Handle("GET /app/workspaces/new", auth.RequireAuthentication(http.HandlerFunc(h.showNew)))
	mux.Handle("POST /app/workspaces/new", auth.RequireAuthentication(http.HandlerFunc(h.create)))

	workspaceMember := middleware.Chain(
		http.HandlerFunc(h.show),
		member.RequireMembership(h.members, h.logger),
	)
	mux.Handle("GET /w/{workspaceSlug}", auth.RequireAuthentication(workspaceMember))

	settings := middleware.Chain(
		http.HandlerFunc(h.showSettings),
		member.RequireMembership(h.members, h.logger),
		member.RequireOwner(),
	)
	mux.Handle("GET /w/{workspaceSlug}/settings", auth.RequireAuthentication(settings))
}

type newPageData struct {
	CSRFToken string
	Form      createFormData
	Errors    CreateErrors
}

type createFormData struct {
	Name string
	Slug string
}

type homePageData struct {
	CSRFToken string
	Workspace Workspace
	User      auth.User
	Role      string
}

type settingsPageData struct {
	CSRFToken string
	Workspace Workspace
	User      auth.User
}

func (h *Handler) showNew(w http.ResponseWriter, r *http.Request) {
	h.renderNew(w, http.StatusOK, newPageData{
		CSRFToken: middleware.CSRFToken(r.Context()),
	})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	ws, fieldErrs, err := h.service.Create(r.Context(), CreateInput{
		Name:      r.FormValue("name"),
		Slug:      r.FormValue("slug"),
		CreatedBy: user.ID,
	})
	if err != nil {
		h.logger.Error("create workspace failed", "err", err, "user_id", user.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if fieldErrs.Any() {
		h.renderNew(w, http.StatusUnprocessableEntity, newPageData{
			CSRFToken: middleware.CSRFToken(r.Context()),
			Form: createFormData{
				Name: strings.TrimSpace(r.FormValue("name")),
				Slug: strings.ToLower(strings.TrimSpace(r.FormValue("slug"))),
			},
			Errors: fieldErrs,
		})
		return
	}

	h.logger.Info("workspace created", "workspace_id", ws.ID, "slug", ws.Slug, "user_id", user.ID)
	http.Redirect(w, r, "/w/"+ws.Slug, http.StatusSeeOther)
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

	h.renderHome(w, http.StatusOK, homePageData{
		CSRFToken: middleware.CSRFToken(r.Context()),
		Workspace: ws,
		User:      user,
		Role:      role,
	})
}

func (h *Handler) showSettings(w http.ResponseWriter, r *http.Request) {
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

	h.renderSettings(w, http.StatusOK, settingsPageData{
		CSRFToken: middleware.CSRFToken(r.Context()),
		Workspace: ws,
		User:      user,
	})
}

func workspaceFromAccessContext(r *http.Request) (Workspace, bool) {
	id, okID := middleware.WorkspaceIDFromContext(r.Context())
	name, okName := middleware.WorkspaceNameFromContext(r.Context())
	slug, okSlug := middleware.WorkspaceSlugFromContext(r.Context())
	if !okID || !okName || !okSlug {
		return Workspace{}, false
	}
	return Workspace{ID: id, Name: name, Slug: slug}, true
}

func (h *Handler) renderNew(w http.ResponseWriter, status int, data newPageData) {
	if err := h.render.Render(w, status, "workspace_new", data); err != nil {
		h.logger.Error("render workspace new failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) renderHome(w http.ResponseWriter, status int, data homePageData) {
	if err := h.render.Render(w, status, "workspace_home", data); err != nil {
		h.logger.Error("render workspace home failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) renderSettings(w http.ResponseWriter, status int, data settingsPageData) {
	if err := h.render.Render(w, status, "workspace_settings", data); err != nil {
		h.logger.Error("render workspace settings failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
