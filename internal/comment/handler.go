package comment

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/issue"
	"github.com/kevin-voss/htmx-go-postgresql/internal/member"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/render"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/request"
	"github.com/kevin-voss/htmx-go-postgresql/internal/project"
)

// Handler serves comment HTTP endpoints.
type Handler struct {
	service  *Service
	issues   *issue.Service
	projects *project.Service
	members  *member.Service
	render   *render.Renderer
	logger   *slog.Logger
}

// NewHandler constructs a comment HTTP handler.
func NewHandler(
	service *Service,
	issues *issue.Service,
	projects *project.Service,
	members *member.Service,
	renderer *render.Renderer,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		service:  service,
		issues:   issues,
		projects: projects,
		members:  members,
		render:   renderer,
		logger:   logger,
	}
}

// Mount registers comment routes on mux (all require authentication + membership).
func (h *Handler) Mount(mux *http.ServeMux) {
	create := middleware.Chain(
		http.HandlerFunc(h.create),
		member.RequireMembership(h.members, h.logger),
		member.RequireCanMutate(),
	)
	mux.Handle("POST /w/{workspaceSlug}/issues/{issueNumber}/comments", auth.RequireAuthentication(create))

	del := middleware.Chain(
		http.HandlerFunc(h.delete),
		member.RequireMembership(h.members, h.logger),
	)
	mux.Handle("DELETE /w/{workspaceSlug}/issues/{issueNumber}/comments/{commentID}", auth.RequireAuthentication(del))
	mux.Handle("POST /w/{workspaceSlug}/issues/{issueNumber}/comments/{commentID}/delete", auth.RequireAuthentication(del))
}

type workspaceAccess struct {
	ID   string
	Name string
	Slug string
}

type commentItemData struct {
	Comment       Comment
	WorkspaceSlug string
	IssueNumber   int
	CSRFToken     string
	CanDelete     bool
}

type commentFormData struct {
	CSRFToken     string
	WorkspaceSlug string
	IssueNumber   int
	Body          string
	Error         string
}

type commentCountData struct {
	Count int
}

type createResponseData struct {
	Item  commentItemData
	Count commentCountData
	Form  commentFormData
}

type deleteResponseData struct {
	CommentID string
	Count     commentCountData
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

	issueNumber, err := strconv.Atoi(r.PathValue("issueNumber"))
	if err != nil || issueNumber < 1 {
		http.NotFound(w, r)
		return
	}

	iss, err := h.issues.GetByWorkspaceAndNumber(r.Context(), ws.ID, issueNumber)
	if err != nil {
		if errors.Is(err, issue.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("get issue for comment failed", "err", err, "workspace_id", ws.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	csrf := middleware.CSRFToken(r.Context())
	body := r.FormValue("body")
	created, fieldErrs, err := h.service.Create(r.Context(), CreateInput{
		IssueID:  iss.ID,
		AuthorID: user.ID,
		Body:     body,
	})
	if err != nil {
		h.logger.Error("create comment failed", "err", err, "issue_id", iss.ID, "user_id", user.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if fieldErrs.Any() {
		form := commentFormData{
			CSRFToken:     csrf,
			WorkspaceSlug: ws.Slug,
			IssueNumber:   issueNumber,
			Body:          strings.TrimSpace(body),
			Error:         fieldErrs.Body,
		}
		if request.IsPartialRequest(r) {
			if err := h.render.RenderFragment(w, http.StatusUnprocessableEntity, "issue_show", "comment_form", form); err != nil {
				h.logger.Error("render comment form errors failed", "err", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		h.redirectToIssue(w, r, ws.Slug, iss)
		return
	}

	count, err := h.service.CountByIssue(r.Context(), iss.ID)
	if err != nil {
		h.logger.Error("count comments failed", "err", err, "issue_id", iss.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	role, _ := middleware.WorkspaceRoleFromContext(r.Context())
	data := createResponseData{
		Item: commentItemData{
			Comment:       created,
			WorkspaceSlug: ws.Slug,
			IssueNumber:   issueNumber,
			CSRFToken:     csrf,
			CanDelete:     created.CanDelete(user.ID, role),
		},
		Count: commentCountData{Count: count},
		Form: commentFormData{
			CSRFToken:     csrf,
			WorkspaceSlug: ws.Slug,
			IssueNumber:   issueNumber,
		},
	}

	if request.IsPartialRequest(r) {
		if err := h.render.RenderFragment(w, http.StatusCreated, "issue_show", "comment_create_response", data); err != nil {
			h.logger.Error("render comment create response failed", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	h.redirectToIssue(w, r, ws.Slug, iss)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
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

	issueNumber, err := strconv.Atoi(r.PathValue("issueNumber"))
	if err != nil || issueNumber < 1 {
		http.NotFound(w, r)
		return
	}

	commentID := strings.TrimSpace(r.PathValue("commentID"))
	if commentID == "" {
		http.NotFound(w, r)
		return
	}

	iss, err := h.issues.GetByWorkspaceAndNumber(r.Context(), ws.ID, issueNumber)
	if err != nil {
		if errors.Is(err, issue.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("get issue for comment delete failed", "err", err, "workspace_id", ws.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	existing, err := h.service.GetByID(r.Context(), commentID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("get comment failed", "err", err, "comment_id", commentID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if existing.IssueID != iss.ID {
		http.NotFound(w, r)
		return
	}

	role, _ := middleware.WorkspaceRoleFromContext(r.Context())
	if err := h.service.Delete(r.Context(), commentID, user.ID, role); err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		h.logger.Error("delete comment failed", "err", err, "comment_id", commentID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if request.IsPartialRequest(r) {
		count, countErr := h.service.CountByIssue(r.Context(), iss.ID)
		if countErr != nil {
			h.logger.Error("count comments failed", "err", countErr, "issue_id", iss.ID)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		data := deleteResponseData{
			CommentID: commentID,
			Count:     commentCountData{Count: count},
		}
		if err := h.render.RenderFragment(w, http.StatusOK, "issue_show", "comment_delete_response", data); err != nil {
			h.logger.Error("render comment delete response failed", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	h.redirectToIssue(w, r, ws.Slug, iss)
}

func (h *Handler) redirectToIssue(w http.ResponseWriter, r *http.Request, workspaceSlug string, iss issue.Issue) {
	p, err := h.projects.GetByID(r.Context(), iss.ProjectID)
	if err != nil {
		http.Redirect(w, r, "/w/"+workspaceSlug+"/issues/"+strconv.Itoa(iss.IssueNumber), http.StatusSeeOther)
		return
	}
	http.Redirect(
		w,
		r,
		"/w/"+workspaceSlug+"/projects/"+p.Slug+"/issues/"+strconv.Itoa(iss.IssueNumber),
		http.StatusSeeOther,
	)
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
