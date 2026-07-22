package issue

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/member"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/render"
	"github.com/kevin-voss/htmx-go-postgresql/internal/project"
)

// Handler serves issue HTTP endpoints.
type Handler struct {
	service  *Service
	projects *project.Service
	members  *member.Service
	render   *render.Renderer
	logger   *slog.Logger
}

// NewHandler constructs an issue HTTP handler.
func NewHandler(
	service *Service,
	projects *project.Service,
	members *member.Service,
	renderer *render.Renderer,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		service:  service,
		projects: projects,
		members:  members,
		render:   renderer,
		logger:   logger,
	}
}

// Mount registers issue routes on mux (all require authentication + membership).
func (h *Handler) Mount(mux *http.ServeMux) {
	list := middleware.Chain(
		http.HandlerFunc(h.list),
		member.RequireMembership(h.members, h.logger),
	)
	mux.Handle("GET /w/{workspaceSlug}/projects/{projectSlug}/issues", auth.RequireAuthentication(list))

	create := middleware.Chain(
		http.HandlerFunc(h.create),
		member.RequireMembership(h.members, h.logger),
		member.RequireCanMutate(),
	)
	mux.Handle("POST /w/{workspaceSlug}/projects/{projectSlug}/issues", auth.RequireAuthentication(create))

	showInProject := middleware.Chain(
		http.HandlerFunc(h.showInProject),
		member.RequireMembership(h.members, h.logger),
	)
	mux.Handle(
		"GET /w/{workspaceSlug}/projects/{projectSlug}/issues/{issueNumber}",
		auth.RequireAuthentication(showInProject),
	)

	show := middleware.Chain(
		http.HandlerFunc(h.show),
		member.RequireMembership(h.members, h.logger),
	)
	mux.Handle("GET /w/{workspaceSlug}/issues/{issueNumber}", auth.RequireAuthentication(show))
}

type listPageData struct {
	CSRFToken     string
	WorkspaceID   string
	WorkspaceName string
	WorkspaceSlug string
	Project       project.Project
	Cards         []cardData
	User          auth.User
	Role          string
	CanCreate     bool
	Form          createFormData
	Errors        CreateErrors
}

type createFormData struct {
	Title       string
	Description string
}

type cardData struct {
	WorkspaceSlug string
	ProjectSlug   string
	Issue         Issue
	DisplayKey    string
	StatusLabel   string
}

type showPageData struct {
	CSRFToken     string
	WorkspaceID   string
	WorkspaceName string
	WorkspaceSlug string
	Project       project.Project
	Issue         Issue
	DisplayKey    string
	StatusLabel   string
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

	p, err := h.projects.GetByWorkspaceAndSlug(r.Context(), ws.ID, r.PathValue("projectSlug"))
	if err != nil {
		if errors.Is(err, project.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("get project failed", "err", err, "workspace_id", ws.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	issues, err := h.service.ListByProject(r.Context(), p.ID)
	if err != nil {
		h.logger.Error("list issues failed", "err", err, "project_id", p.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.renderList(w, http.StatusOK, listPageData{
		CSRFToken:     middleware.CSRFToken(r.Context()),
		WorkspaceID:   ws.ID,
		WorkspaceName: ws.Name,
		WorkspaceSlug: ws.Slug,
		Project:       p,
		Cards:         cardsFor(ws.Slug, p.Slug, issues),
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

	p, err := h.projects.GetByWorkspaceAndSlug(r.Context(), ws.ID, r.PathValue("projectSlug"))
	if err != nil {
		if errors.Is(err, project.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("get project failed", "err", err, "workspace_id", ws.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	issue, fieldErrs, err := h.service.Create(r.Context(), CreateInput{
		ProjectID:   p.ID,
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		CreatedBy:   user.ID,
	})
	if err != nil {
		h.logger.Error("create issue failed", "err", err, "project_id", p.ID, "user_id", user.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if fieldErrs.Any() {
		issues, listErr := h.service.ListByProject(r.Context(), p.ID)
		if listErr != nil {
			h.logger.Error("list issues failed", "err", listErr, "project_id", p.ID)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		h.renderList(w, http.StatusUnprocessableEntity, listPageData{
			CSRFToken:     middleware.CSRFToken(r.Context()),
			WorkspaceID:   ws.ID,
			WorkspaceName: ws.Name,
			WorkspaceSlug: ws.Slug,
			Project:       p,
			Cards:         cardsFor(ws.Slug, p.Slug, issues),
			User:          user,
			Role:          role,
			CanCreate:     true,
			Form: createFormData{
				Title:       strings.TrimSpace(r.FormValue("title")),
				Description: strings.TrimSpace(r.FormValue("description")),
			},
			Errors: fieldErrs,
		})
		return
	}

	h.logger.Info(
		"issue created",
		"issue_id", issue.ID,
		"issue_number", issue.IssueNumber,
		"project_id", p.ID,
		"user_id", user.ID,
	)
	http.Redirect(
		w,
		r,
		"/w/"+ws.Slug+"/projects/"+p.Slug+"/issues/"+strconv.Itoa(issue.IssueNumber),
		http.StatusSeeOther,
	)
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

	issueNumber, err := strconv.Atoi(r.PathValue("issueNumber"))
	if err != nil || issueNumber < 1 {
		http.NotFound(w, r)
		return
	}

	issue, err := h.service.GetByWorkspaceAndNumber(r.Context(), ws.ID, issueNumber)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("get issue failed", "err", err, "workspace_id", ws.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	p, err := h.projects.GetByID(r.Context(), issue.ProjectID)
	if err != nil {
		if errors.Is(err, project.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("get project for issue failed", "err", err, "project_id", issue.ProjectID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if p.WorkspaceID != ws.ID {
		http.NotFound(w, r)
		return
	}

	h.renderShow(w, showPageData{
		CSRFToken:     middleware.CSRFToken(r.Context()),
		WorkspaceID:   ws.ID,
		WorkspaceName: ws.Name,
		WorkspaceSlug: ws.Slug,
		Project:       p,
		Issue:         issue,
		DisplayKey:    DisplayKey(p.Slug, issue.IssueNumber),
		StatusLabel:   StatusLabel(issue.Status),
		User:          user,
		Role:          role,
	})
}

func (h *Handler) showInProject(w http.ResponseWriter, r *http.Request) {
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

	issueNumber, err := strconv.Atoi(r.PathValue("issueNumber"))
	if err != nil || issueNumber < 1 {
		http.NotFound(w, r)
		return
	}

	p, err := h.projects.GetByWorkspaceAndSlug(r.Context(), ws.ID, r.PathValue("projectSlug"))
	if err != nil {
		if errors.Is(err, project.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("get project failed", "err", err, "workspace_id", ws.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	issue, err := h.service.GetByProjectAndNumber(r.Context(), p.ID, issueNumber)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("get issue failed", "err", err, "project_id", p.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.renderShow(w, showPageData{
		CSRFToken:     middleware.CSRFToken(r.Context()),
		WorkspaceID:   ws.ID,
		WorkspaceName: ws.Name,
		WorkspaceSlug: ws.Slug,
		Project:       p,
		Issue:         issue,
		DisplayKey:    DisplayKey(p.Slug, issue.IssueNumber),
		StatusLabel:   StatusLabel(issue.Status),
		User:          user,
		Role:          role,
	})
}

func cardsFor(workspaceSlug, projectSlug string, issues []Issue) []cardData {
	cards := make([]cardData, 0, len(issues))
	for _, issue := range issues {
		cards = append(cards, cardData{
			WorkspaceSlug: workspaceSlug,
			ProjectSlug:   projectSlug,
			Issue:         issue,
			DisplayKey:    DisplayKey(projectSlug, issue.IssueNumber),
			StatusLabel:   StatusLabel(issue.Status),
		})
	}
	return cards
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
	if err := h.render.Render(w, status, "issue_list", data); err != nil {
		h.logger.Error("render issue list failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) renderShow(w http.ResponseWriter, data showPageData) {
	if err := h.render.Render(w, http.StatusOK, "issue_show", data); err != nil {
		h.logger.Error("render issue show failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
