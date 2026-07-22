package workspace

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/member"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/render"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/ui"
	"github.com/kevin-voss/htmx-go-postgresql/internal/project"
)

// Handler serves workspace HTTP endpoints.
type Handler struct {
	service  *Service
	members  *member.Service
	projects *project.Service
	render   *render.Renderer
	logger   *slog.Logger
}

// NewHandler constructs a workspace HTTP handler.
func NewHandler(
	service *Service,
	members *member.Service,
	projects *project.Service,
	renderer *render.Renderer,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		service:  service,
		members:  members,
		projects: projects,
		render:   renderer,
		logger:   logger,
	}
}

// Mount registers workspace routes on mux (all require authentication).
func (h *Handler) Mount(mux *http.ServeMux) {
	mux.Handle("GET /app/onboarding", auth.RequireAuthentication(http.HandlerFunc(h.showOnboarding)))
	mux.Handle("POST /app/onboarding", auth.RequireAuthentication(http.HandlerFunc(h.completeOnboarding)))
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
	Chrome    ui.Chrome
}

type createFormData struct {
	Name string
	Slug string
}

type onboardingPageData struct {
	CSRFToken string
	Form      onboardingFormData
	Errors    OnboardErrors
}

type onboardingFormData struct {
	Name        string
	Slug        string
	ProjectName string
}

type homePageData struct {
	CSRFToken string
	Workspace Workspace
	Projects  []project.Project
	User      auth.User
	Role      string
	Chrome    ui.Chrome
}

type settingsPageData struct {
	CSRFToken string
	Workspace Workspace
	User      auth.User
	Chrome    ui.Chrome
}

func (h *Handler) showOnboarding(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	has, err := h.members.HasAnyMembership(r.Context(), user.ID)
	if err != nil {
		h.logger.Error("check memberships failed", "err", err, "user_id", user.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if has {
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}

	h.renderOnboarding(w, http.StatusOK, onboardingPageData{
		CSRFToken: middleware.CSRFToken(r.Context()),
	})
}

func (h *Handler) completeOnboarding(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	has, err := h.members.HasAnyMembership(r.Context(), user.ID)
	if err != nil {
		h.logger.Error("check memberships failed", "err", err, "user_id", user.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if has {
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	result, fieldErrs, err := h.service.Onboard(r.Context(), OnboardInput{
		Name:        r.FormValue("name"),
		Slug:        r.FormValue("slug"),
		ProjectName: r.FormValue("project_name"),
		CreatedBy:   user.ID,
	})
	if err != nil {
		h.logger.Error("onboarding failed", "err", err, "user_id", user.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if fieldErrs.Any() {
		h.renderOnboarding(w, http.StatusUnprocessableEntity, onboardingPageData{
			CSRFToken: middleware.CSRFToken(r.Context()),
			Form: onboardingFormData{
				Name:        strings.TrimSpace(r.FormValue("name")),
				Slug:        strings.ToLower(strings.TrimSpace(r.FormValue("slug"))),
				ProjectName: strings.TrimSpace(r.FormValue("project_name")),
			},
			Errors: fieldErrs,
		})
		return
	}

	h.logger.Info(
		"onboarding completed",
		"workspace_id", result.Workspace.ID,
		"slug", result.Workspace.Slug,
		"project_id", result.ProjectID,
		"user_id", user.ID,
	)
	http.Redirect(
		w,
		r,
		"/w/"+result.Workspace.Slug+"/projects/"+result.ProjectSlug,
		http.StatusSeeOther,
	)
}

func (h *Handler) showNew(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	csrf := middleware.CSRFToken(r.Context())
	h.renderNew(w, http.StatusOK, newPageData{
		CSRFToken: csrf,
		Chrome: ui.App(user.DisplayName, csrf,
			ui.Crumb{Label: "App", Href: "/app"},
			ui.Crumb{Label: "New workspace"},
		),
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
		csrf := middleware.CSRFToken(r.Context())
		h.renderNew(w, http.StatusUnprocessableEntity, newPageData{
			CSRFToken: csrf,
			Form: createFormData{
				Name: strings.TrimSpace(r.FormValue("name")),
				Slug: strings.ToLower(strings.TrimSpace(r.FormValue("slug"))),
			},
			Errors: fieldErrs,
			Chrome: ui.App(user.DisplayName, csrf,
				ui.Crumb{Label: "App", Href: "/app"},
				ui.Crumb{Label: "New workspace"},
			),
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
	csrf := middleware.CSRFToken(r.Context())

	var projects []project.Project
	if h.projects != nil {
		listed, err := h.projects.ListByWorkspace(r.Context(), ws.ID)
		if err != nil {
			h.logger.Error("list projects failed", "err", err, "workspace_id", ws.ID)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		projects = listed
	}

	h.renderHome(w, http.StatusOK, homePageData{
		CSRFToken: csrf,
		Workspace: ws,
		Projects:  projects,
		User:      user,
		Role:      role,
		Chrome: ui.Workspace(user.DisplayName, csrf, ws.Name, ws.Slug, role, "",
			ui.Crumb{Label: "App", Href: "/app"},
			ui.Crumb{Label: ws.Name},
		),
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
	role, _ := middleware.WorkspaceRoleFromContext(r.Context())
	csrf := middleware.CSRFToken(r.Context())

	h.renderSettings(w, http.StatusOK, settingsPageData{
		CSRFToken: csrf,
		Workspace: ws,
		User:      user,
		Chrome: ui.Workspace(user.DisplayName, csrf, ws.Name, ws.Slug, role, ui.NavSettings,
			ui.Crumb{Label: "App", Href: "/app"},
			ui.Crumb{Label: ws.Name, Href: "/w/" + ws.Slug},
			ui.Crumb{Label: "Settings"},
		),
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

func (h *Handler) renderOnboarding(w http.ResponseWriter, status int, data onboardingPageData) {
	if err := h.render.Render(w, status, "onboarding", data); err != nil {
		h.logger.Error("render onboarding failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
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
