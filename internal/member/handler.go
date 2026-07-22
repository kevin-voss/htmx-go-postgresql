package member

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/mail"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/render"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/ui"
)

// Handler serves membership and invitation HTTP endpoints.
type Handler struct {
	service *Service
	mailer  mail.Sender
	render  *render.Renderer
	logger  *slog.Logger
}

// NewHandler constructs a member HTTP handler.
func NewHandler(service *Service, mailer mail.Sender, renderer *render.Renderer, logger *slog.Logger) *Handler {
	if mailer == nil {
		mailer = mail.NopMailer{}
	}
	return &Handler{
		service: service,
		mailer:  mailer,
		render:  renderer,
		logger:  logger,
	}
}

// Mount registers member and invitation routes on mux.
func (h *Handler) Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /invites/{token}", h.showInvite)

	members := middleware.Chain(
		http.HandlerFunc(h.showMembers),
		RequireMembership(h.service, h.logger),
	)
	mux.Handle("GET /w/{workspaceSlug}/members", auth.RequireAuthentication(members))

	invite := middleware.Chain(
		http.HandlerFunc(h.createInvite),
		RequireMembership(h.service, h.logger),
		RequireOwner(),
	)
	mux.Handle("POST /w/{workspaceSlug}/members/invites", auth.RequireAuthentication(invite))

	changeRole := middleware.Chain(
		http.HandlerFunc(h.changeRole),
		RequireMembership(h.service, h.logger),
		RequireOwner(),
	)
	mux.Handle("POST /w/{workspaceSlug}/members/{userID}/role", auth.RequireAuthentication(changeRole))

	remove := middleware.Chain(
		http.HandlerFunc(h.removeMember),
		RequireMembership(h.service, h.logger),
		RequireOwner(),
	)
	mux.Handle("POST /w/{workspaceSlug}/members/{userID}/remove", auth.RequireAuthentication(remove))
}

type membersPageData struct {
	CSRFToken     string
	WorkspaceID   string
	WorkspaceName string
	WorkspaceSlug string
	User          auth.User
	Role          string
	Members       []MemberView
	IsOwner       bool
	Form          inviteFormData
	Errors        InviteErrors
	Flash         string
	Chrome        ui.Chrome
}

type inviteFormData struct {
	Email string
	Role  string
}

type inviteAcceptPageData struct {
	CSRFToken string
	Token     string
	Invite    Invitation
	User      *auth.User
	Error     string
	Mismatch  bool
	Invalid   bool
	Already   bool
	Accepted  bool
}

func (h *Handler) showMembers(w http.ResponseWriter, r *http.Request) {
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

	members, err := h.service.ListMembers(r.Context(), workspaceID)
	if err != nil {
		h.logger.Error("list members failed", "err", err, "workspace_id", workspaceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	csrf := middleware.CSRFToken(r.Context())
	h.renderMembers(w, http.StatusOK, membersPageData{
		CSRFToken:     csrf,
		WorkspaceID:   workspaceID,
		WorkspaceName: workspaceName,
		WorkspaceSlug: workspaceSlug,
		User:          user,
		Role:          role,
		Members:       members,
		IsOwner:       Role(role).CanInvite(),
		Flash:         strings.TrimSpace(r.URL.Query().Get("flash")),
		Chrome:        membersChrome(user.DisplayName, csrf, workspaceName, workspaceSlug, role),
	})
}

func (h *Handler) createInvite(w http.ResponseWriter, r *http.Request) {
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

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	inviteRole := Role(strings.TrimSpace(r.FormValue("role")))
	rawToken, fieldErrs, err := h.service.CreateInvitation(r.Context(), CreateInvitationInput{
		WorkspaceID:   workspaceID,
		WorkspaceName: workspaceName,
		Email:         r.FormValue("email"),
		Role:          inviteRole,
		InvitedBy:     user.ID,
	})
	if err != nil {
		h.logger.Error("create invitation failed", "err", err, "workspace_id", workspaceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if fieldErrs.Any() {
		members, listErr := h.service.ListMembers(r.Context(), workspaceID)
		if listErr != nil {
			h.logger.Error("list members failed", "err", listErr, "workspace_id", workspaceID)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		csrf := middleware.CSRFToken(r.Context())
		h.renderMembers(w, http.StatusUnprocessableEntity, membersPageData{
			CSRFToken:     csrf,
			WorkspaceID:   workspaceID,
			WorkspaceName: workspaceName,
			WorkspaceSlug: workspaceSlug,
			User:          user,
			Role:          role,
			Members:       members,
			IsOwner:       true,
			Form: inviteFormData{
				Email: strings.ToLower(strings.TrimSpace(r.FormValue("email"))),
				Role:  string(inviteRole),
			},
			Errors: fieldErrs,
			Chrome: membersChrome(user.DisplayName, csrf, workspaceName, workspaceSlug, role),
		})
		return
	}

	inviteURL := requestBaseURL(r) + "/invites/" + url.PathEscape(rawToken)
	if err := h.mailer.Send(mail.Message{
		To:      strings.ToLower(strings.TrimSpace(r.FormValue("email"))),
		Subject: "You're invited to " + workspaceName + " on Forgeboard",
		Body:    invitationEmailBody(workspaceName, inviteURL),
	}); err != nil {
		h.logger.Error("send invitation email failed", "err", err, "workspace_id", workspaceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("invitation sent", "workspace_id", workspaceID, "invited_by", user.ID)
	http.Redirect(w, r, "/w/"+workspaceSlug+"/members?flash=invited", http.StatusSeeOther)
}

func (h *Handler) changeRole(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := middleware.WorkspaceIDFromContext(r.Context())
	if !ok {
		http.NotFound(w, r)
		return
	}
	workspaceSlug, _ := middleware.WorkspaceSlugFromContext(r.Context())
	targetUserID := strings.TrimSpace(r.PathValue("userID"))

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	newRole := Role(strings.TrimSpace(r.FormValue("role")))
	if err := h.service.ChangeRole(r.Context(), workspaceID, targetUserID, newRole); err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		h.logger.Error("change role failed", "err", err, "workspace_id", workspaceID, "user_id", targetUserID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/w/"+workspaceSlug+"/members?flash=role", http.StatusSeeOther)
}

func (h *Handler) removeMember(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := middleware.WorkspaceIDFromContext(r.Context())
	if !ok {
		http.NotFound(w, r)
		return
	}
	workspaceSlug, _ := middleware.WorkspaceSlugFromContext(r.Context())
	targetUserID := strings.TrimSpace(r.PathValue("userID"))

	if err := h.service.RemoveMember(r.Context(), workspaceID, targetUserID); err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		h.logger.Error("remove member failed", "err", err, "workspace_id", workspaceID, "user_id", targetUserID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/w/"+workspaceSlug+"/members?flash=removed", http.StatusSeeOther)
}

func (h *Handler) showInvite(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.PathValue("token"))
	inv, err := h.service.LoadInvitation(r.Context(), token)
	if err != nil {
		if errors.Is(err, ErrInvalidInvitation) {
			h.renderInviteAccept(w, http.StatusBadRequest, inviteAcceptPageData{
				Token:   token,
				Invalid: true,
				Error:   "This invitation link is invalid or has expired.",
			})
			return
		}
		h.logger.Error("load invitation failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	user, loggedIn := auth.UserFromContext(r.Context())
	data := inviteAcceptPageData{
		CSRFToken: middleware.CSRFToken(r.Context()),
		Token:     token,
		Invite:    inv,
	}
	if loggedIn {
		data.User = &user
	}

	if !loggedIn {
		h.renderInviteAccept(w, http.StatusOK, data)
		return
	}

	result, err := h.service.AcceptInvitation(r.Context(), token, user.ID, user.Email)
	if err != nil {
		if errors.Is(err, ErrEmailMismatch) {
			data.Mismatch = true
			data.Error = "This invitation was sent to " + inv.Email + ". Sign in with that email to accept."
			h.renderInviteAccept(w, http.StatusForbidden, data)
			return
		}
		if errors.Is(err, ErrInvalidInvitation) {
			h.renderInviteAccept(w, http.StatusBadRequest, inviteAcceptPageData{
				Token:   token,
				Invalid: true,
				Error:   "This invitation link is invalid or has expired.",
			})
			return
		}
		h.logger.Error("accept invitation failed", "err", err, "user_id", user.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if result.Already {
		data.Already = true
		h.renderInviteAccept(w, http.StatusOK, data)
		return
	}

	http.Redirect(w, r, "/w/"+result.Invitation.WorkspaceSlug, http.StatusSeeOther)
}

func membersChrome(displayName, csrf, name, slug, role string) ui.Chrome {
	return ui.Workspace(displayName, csrf, name, slug, role, ui.NavMembers,
		ui.Crumb{Label: "App", Href: "/app"},
		ui.Crumb{Label: name, Href: "/w/" + slug},
		ui.Crumb{Label: "Members"},
	)
}

func (h *Handler) renderMembers(w http.ResponseWriter, status int, data membersPageData) {
	if err := h.render.Render(w, status, "members", data); err != nil {
		h.logger.Error("render members failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) renderInviteAccept(w http.ResponseWriter, status int, data inviteAcceptPageData) {
	if err := h.render.Render(w, status, "invite_accept", data); err != nil {
		h.logger.Error("render invite accept failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func requestBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		host = "localhost:8080"
	}
	return scheme + "://" + host
}

func invitationEmailBody(workspaceName, inviteURL string) string {
	name := strings.TrimSpace(workspaceName)
	if name == "" {
		name = "a Forgeboard workspace"
	}
	return fmt.Sprintf(
		"You've been invited to join %s on Forgeboard.\n\nAccept the invitation:\n\n%s\n\nThis link expires in 7 days.\n",
		name,
		inviteURL,
	)
}
