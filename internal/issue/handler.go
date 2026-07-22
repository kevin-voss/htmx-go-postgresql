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

	updateStatus := middleware.Chain(
		http.HandlerFunc(h.updateStatus),
		member.RequireMembership(h.members, h.logger),
		member.RequireCanMutate(),
	)
	mux.Handle("PATCH /w/{workspaceSlug}/issues/{issueNumber}/status", auth.RequireAuthentication(updateStatus))
	mux.Handle("POST /w/{workspaceSlug}/issues/{issueNumber}/status", auth.RequireAuthentication(updateStatus))

	updatePriority := middleware.Chain(
		http.HandlerFunc(h.updatePriority),
		member.RequireMembership(h.members, h.logger),
		member.RequireCanMutate(),
	)
	mux.Handle("PATCH /w/{workspaceSlug}/issues/{issueNumber}/priority", auth.RequireAuthentication(updatePriority))
	mux.Handle("POST /w/{workspaceSlug}/issues/{issueNumber}/priority", auth.RequireAuthentication(updatePriority))

	updateAssignee := middleware.Chain(
		http.HandlerFunc(h.updateAssignee),
		member.RequireMembership(h.members, h.logger),
		member.RequireCanMutate(),
	)
	mux.Handle("PATCH /w/{workspaceSlug}/issues/{issueNumber}/assignee", auth.RequireAuthentication(updateAssignee))
	mux.Handle("POST /w/{workspaceSlug}/issues/{issueNumber}/assignee", auth.RequireAuthentication(updateAssignee))

	archive := middleware.Chain(
		http.HandlerFunc(h.archive),
		member.RequireMembership(h.members, h.logger),
		member.RequireCanMutate(),
	)
	mux.Handle("POST /w/{workspaceSlug}/issues/{issueNumber}/archive", auth.RequireAuthentication(archive))
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
	CanEdit       bool
	Form          createFormData
	Errors        CreateErrors
	Statuses      []optionData
	Priorities    []optionData
	Members       []memberOption
}

type createFormData struct {
	Title       string
	Description string
}

type cardData struct {
	WorkspaceSlug  string
	ProjectSlug    string
	Issue          Issue
	DisplayKey     string
	StatusLabel    string
	PriorityLabel  string
	AssigneeLabel  string
	CSRFToken      string
	CanEdit        bool
	Statuses       []optionData
	Priorities     []optionData
	Members        []memberOption
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
	PriorityLabel string
	AssigneeLabel string
	User          auth.User
	Role          string
	CanEdit       bool
	Statuses      []optionData
	Priorities    []optionData
	Members       []memberOption
}

type optionData struct {
	Value string
	Label string
}

type memberOption struct {
	UserID      string
	DisplayName string
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
	canEdit := member.Role(role).CanMutate()

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

	members, memberOpts, err := h.loadMembers(r, ws.ID)
	if err != nil {
		h.logger.Error("list members failed", "err", err, "workspace_id", ws.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	csrf := middleware.CSRFToken(r.Context())
	h.renderList(w, http.StatusOK, listPageData{
		CSRFToken:     csrf,
		WorkspaceID:   ws.ID,
		WorkspaceName: ws.Name,
		WorkspaceSlug: ws.Slug,
		Project:       p,
		Cards:         cardsFor(ws.Slug, p.Slug, issues, csrf, canEdit, members, memberOpts),
		User:          user,
		Role:          role,
		CanCreate:     canEdit,
		CanEdit:       canEdit,
		Statuses:      statusOptions(),
		Priorities:    priorityOptions(),
		Members:       memberOpts,
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
		members, memberOpts, memErr := h.loadMembers(r, ws.ID)
		if memErr != nil {
			h.logger.Error("list members failed", "err", memErr, "workspace_id", ws.ID)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		csrf := middleware.CSRFToken(r.Context())
		h.renderList(w, http.StatusUnprocessableEntity, listPageData{
			CSRFToken:     csrf,
			WorkspaceID:   ws.ID,
			WorkspaceName: ws.Name,
			WorkspaceSlug: ws.Slug,
			Project:       p,
			Cards:         cardsFor(ws.Slug, p.Slug, issues, csrf, true, members, memberOpts),
			User:          user,
			Role:          role,
			CanCreate:     true,
			CanEdit:       true,
			Form: createFormData{
				Title:       strings.TrimSpace(r.FormValue("title")),
				Description: strings.TrimSpace(r.FormValue("description")),
			},
			Errors:     fieldErrs,
			Statuses:   statusOptions(),
			Priorities: priorityOptions(),
			Members:    memberOpts,
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

	h.renderShowPage(w, r, ws, p, issue, user, role)
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

	h.renderShowPage(w, r, ws, p, issue, user, role)
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	ws, issueNumber, ok := h.parseWorkspaceIssue(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	issue, err := h.service.UpdateStatus(r.Context(), ws.ID, issueNumber, r.FormValue("status"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, ErrInvalidStatus) {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		h.logger.Error("update status failed", "err", err, "workspace_id", ws.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.redirectAfterMutation(w, r, ws.Slug, issue)
}

func (h *Handler) updatePriority(w http.ResponseWriter, r *http.Request) {
	ws, issueNumber, ok := h.parseWorkspaceIssue(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	issue, err := h.service.UpdatePriority(r.Context(), ws.ID, issueNumber, r.FormValue("priority"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, ErrInvalidPriority) {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		h.logger.Error("update priority failed", "err", err, "workspace_id", ws.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.redirectAfterMutation(w, r, ws.Slug, issue)
}

func (h *Handler) updateAssignee(w http.ResponseWriter, r *http.Request) {
	ws, issueNumber, ok := h.parseWorkspaceIssue(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	issue, err := h.service.UpdateAssignee(r.Context(), ws.ID, issueNumber, r.FormValue("assignee_id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, ErrInvalidAssignee) {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		h.logger.Error("update assignee failed", "err", err, "workspace_id", ws.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.redirectAfterMutation(w, r, ws.Slug, issue)
}

func (h *Handler) archive(w http.ResponseWriter, r *http.Request) {
	ws, issueNumber, ok := h.parseWorkspaceIssue(w, r)
	if !ok {
		return
	}

	issue, err := h.service.Archive(r.Context(), ws.ID, issueNumber)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("archive issue failed", "err", err, "workspace_id", ws.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	p, err := h.projects.GetByID(r.Context(), issue.ProjectID)
	if err != nil {
		http.Redirect(w, r, "/w/"+ws.Slug+"/projects", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/w/"+ws.Slug+"/projects/"+p.Slug+"/issues", http.StatusSeeOther)
}

func (h *Handler) parseWorkspaceIssue(w http.ResponseWriter, r *http.Request) (workspaceAccess, int, bool) {
	ws, ok := workspaceFromAccessContext(r)
	if !ok {
		http.NotFound(w, r)
		return workspaceAccess{}, 0, false
	}
	issueNumber, err := strconv.Atoi(r.PathValue("issueNumber"))
	if err != nil || issueNumber < 1 {
		http.NotFound(w, r)
		return workspaceAccess{}, 0, false
	}
	return ws, issueNumber, true
}

func (h *Handler) redirectAfterMutation(w http.ResponseWriter, r *http.Request, workspaceSlug string, issue Issue) {
	redirectTo := strings.TrimSpace(r.FormValue("redirect_to"))
	if redirectTo != "" && strings.HasPrefix(redirectTo, "/w/") {
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return
	}
	p, err := h.projects.GetByID(r.Context(), issue.ProjectID)
	if err != nil {
		http.Redirect(w, r, "/w/"+workspaceSlug+"/issues/"+strconv.Itoa(issue.IssueNumber), http.StatusSeeOther)
		return
	}
	http.Redirect(
		w,
		r,
		"/w/"+workspaceSlug+"/projects/"+p.Slug+"/issues/"+strconv.Itoa(issue.IssueNumber),
		http.StatusSeeOther,
	)
}

func (h *Handler) renderShowPage(
	w http.ResponseWriter,
	r *http.Request,
	ws workspaceAccess,
	p project.Project,
	issue Issue,
	user auth.User,
	role string,
) {
	members, memberOpts, err := h.loadMembers(r, ws.ID)
	if err != nil {
		h.logger.Error("list members failed", "err", err, "workspace_id", ws.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	canEdit := member.Role(role).CanMutate()
	h.renderShow(w, showPageData{
		CSRFToken:     middleware.CSRFToken(r.Context()),
		WorkspaceID:   ws.ID,
		WorkspaceName: ws.Name,
		WorkspaceSlug: ws.Slug,
		Project:       p,
		Issue:         issue,
		DisplayKey:    DisplayKey(p.Slug, issue.IssueNumber),
		StatusLabel:   StatusLabel(issue.Status),
		PriorityLabel: PriorityLabel(issue.Priority),
		AssigneeLabel: assigneeLabel(issue.AssigneeID, members),
		User:          user,
		Role:          role,
		CanEdit:       canEdit,
		Statuses:      statusOptions(),
		Priorities:    priorityOptions(),
		Members:       memberOpts,
	})
}

func (h *Handler) loadMembers(r *http.Request, workspaceID string) ([]member.MemberView, []memberOption, error) {
	members, err := h.members.ListMembers(r.Context(), workspaceID)
	if err != nil {
		return nil, nil, err
	}
	opts := make([]memberOption, 0, len(members))
	for _, m := range members {
		opts = append(opts, memberOption{UserID: m.UserID, DisplayName: m.DisplayName})
	}
	return members, opts, nil
}

func cardsFor(
	workspaceSlug, projectSlug string,
	issues []Issue,
	csrf string,
	canEdit bool,
	members []member.MemberView,
	memberOpts []memberOption,
) []cardData {
	statuses := statusOptions()
	priorities := priorityOptions()
	cards := make([]cardData, 0, len(issues))
	for _, issue := range issues {
		cards = append(cards, cardData{
			WorkspaceSlug: workspaceSlug,
			ProjectSlug:   projectSlug,
			Issue:         issue,
			DisplayKey:    DisplayKey(projectSlug, issue.IssueNumber),
			StatusLabel:   StatusLabel(issue.Status),
			PriorityLabel: PriorityLabel(issue.Priority),
			AssigneeLabel: assigneeLabel(issue.AssigneeID, members),
			CSRFToken:     csrf,
			CanEdit:       canEdit,
			Statuses:      statuses,
			Priorities:    priorities,
			Members:       memberOpts,
		})
	}
	return cards
}

func statusOptions() []optionData {
	opts := make([]optionData, 0, len(Statuses()))
	for _, s := range Statuses() {
		opts = append(opts, optionData{Value: s, Label: StatusLabel(s)})
	}
	return opts
}

func priorityOptions() []optionData {
	opts := make([]optionData, 0, len(Priorities()))
	for _, p := range Priorities() {
		opts = append(opts, optionData{Value: p, Label: PriorityLabel(p)})
	}
	return opts
}

func assigneeLabel(assigneeID string, members []member.MemberView) string {
	if assigneeID == "" {
		return "Unassigned"
	}
	for _, m := range members {
		if m.UserID == assigneeID {
			return m.DisplayName
		}
	}
	return "Unknown member"
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
