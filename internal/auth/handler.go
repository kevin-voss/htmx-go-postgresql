package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kevin-voss/htmx-go-postgresql/internal/mail"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/render"
)

const (
	invalidCredentialsMessage = "Invalid email or password."
	forgotPasswordAckMessage  = "If an account exists for that email, we sent password reset instructions."
	loginRateLimitedMessage   = "Too many login attempts. Please try again later."
)

// Handler serves auth HTTP endpoints.
type Handler struct {
	service      *Service
	mailer       mail.Sender
	render       *render.Renderer
	logger       *slog.Logger
	cookieSecure bool
	loginLimit   *LoginRateLimiter
}

// NewHandler constructs an auth HTTP handler.
func NewHandler(service *Service, mailer mail.Sender, renderer *render.Renderer, logger *slog.Logger, cookieSecure bool) *Handler {
	if mailer == nil {
		mailer = mail.NopMailer{}
	}
	return &Handler{
		service:      service,
		mailer:       mailer,
		render:       renderer,
		logger:       logger,
		cookieSecure: cookieSecure,
		loginLimit:   NewLoginRateLimiter(),
	}
}

// Mount registers auth routes on mux.
func (h *Handler) Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /register", h.showRegister)
	mux.HandleFunc("POST /register", h.register)
	mux.HandleFunc("GET /login", h.showLogin)
	mux.HandleFunc("POST /login", h.login)
	mux.HandleFunc("POST /logout", h.logout)
	mux.HandleFunc("GET /verify-email", h.verifyEmail)
	mux.HandleFunc("GET /forgot-password", h.showForgotPassword)
	mux.HandleFunc("POST /forgot-password", h.forgotPassword)
	mux.HandleFunc("GET /reset-password/{token}", h.showResetPassword)
	mux.HandleFunc("POST /reset-password/{token}", h.resetPassword)
	mux.Handle("GET /account/sessions", RequireAuthentication(http.HandlerFunc(h.showSessions)))
	mux.Handle("POST /account/sessions/{id}/revoke", RequireAuthentication(http.HandlerFunc(h.revokeSession)))
	mux.Handle("DELETE /account/sessions/{id}", RequireAuthentication(http.HandlerFunc(h.revokeSession)))
}

// LoadSessionMiddleware returns middleware that populates session/user context.
func (h *Handler) LoadSessionMiddleware() middleware.Middleware {
	return LoadSession(h.service, h.cookieSecure, h.logger)
}

type registerPageData struct {
	CSRFToken string
	Form      registerFormData
	Errors    RegisterErrors
}

type registerFormData struct {
	DisplayName string
	Email       string
	AcceptTerms bool
}

type loginPageData struct {
	CSRFToken string
	Form      loginFormData
	Error     string
}

type loginFormData struct {
	Email string
}

type forgotPasswordPageData struct {
	CSRFToken string
	Form      forgotPasswordFormData
	Message   string
}

type forgotPasswordFormData struct {
	Email string
}

type resetPasswordPageData struct {
	CSRFToken string
	Token     string
	Errors    ResetPasswordErrors
	Success   bool
}

type sessionsPageData struct {
	CSRFToken string
	Sessions  []sessionRow
}

type sessionRow struct {
	ID         string
	UserAgent  string
	IPAddress  string
	LastSeenAt string
	CreatedAt  string
	Current    bool
}

func (h *Handler) showRegister(w http.ResponseWriter, r *http.Request) {
	h.renderRegister(w, http.StatusOK, registerPageData{
		CSRFToken: middleware.CSRFToken(r.Context()),
	})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	in := RegisterInput{
		DisplayName:          r.FormValue("display_name"),
		Email:                r.FormValue("email"),
		Password:             r.FormValue("password"),
		PasswordConfirmation: r.FormValue("password_confirmation"),
		AcceptTerms:          acceptedTerms(r.FormValue("terms")),
	}

	user, fieldErrs, err := h.service.Register(r.Context(), in)
	if err != nil {
		h.logger.Error("register failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if fieldErrs.Any() {
		h.renderRegister(w, http.StatusUnprocessableEntity, registerPageData{
			CSRFToken: middleware.CSRFToken(r.Context()),
			Form: registerFormData{
				DisplayName: strings.TrimSpace(in.DisplayName),
				Email:       strings.ToLower(strings.TrimSpace(in.Email)),
				AcceptTerms: in.AcceptTerms,
			},
			Errors: fieldErrs,
		})
		return
	}

	verifyToken, err := h.service.CreateEmailVerificationToken(r.Context(), user.ID)
	if err != nil {
		h.logger.Error("create verification token failed", "err", err, "user_id", user.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	verifyURL := requestBaseURL(r) + "/verify-email?token=" + url.QueryEscape(verifyToken)
	if err := h.mailer.Send(mail.Message{
		To:      user.Email,
		Subject: "Verify your Forgeboard email",
		Body:    verificationEmailBody(user.DisplayName, verifyURL),
	}); err != nil {
		h.logger.Error("send verification email failed", "err", err, "user_id", user.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, rawToken, err := h.service.CreateSession(r.Context(), CreateSessionInput{
		UserID:    user.ID,
		UserAgent: r.UserAgent(),
		IPAddress: ClientIP(r),
	})
	if err != nil {
		h.logger.Error("create session after register failed", "err", err, "user_id", user.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	SetSessionCookie(w, rawToken, h.cookieSecure)
	h.logger.Info("user registered", "user_id", user.ID, "email", user.Email)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) verifyEmail(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if err := h.service.VerifyEmail(r.Context(), token); err != nil {
		if errors.Is(err, ErrInvalidVerificationToken) {
			h.renderVerifyEmailError(w, http.StatusBadRequest)
			return
		}
		h.logger.Error("verify email failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	h.renderVerifyEmailSuccess(w, http.StatusOK)
}

func (h *Handler) showLogin(w http.ResponseWriter, r *http.Request) {
	h.renderLogin(w, http.StatusOK, loginPageData{
		CSRFToken: middleware.CSRFToken(r.Context()),
	})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	ip := ClientIP(r)
	if !h.loginLimit.Allow(ip) {
		h.renderLogin(w, http.StatusTooManyRequests, loginPageData{
			CSRFToken: middleware.CSRFToken(r.Context()),
			Form:      loginFormData{Email: strings.ToLower(strings.TrimSpace(r.FormValue("email")))},
			Error:     loginRateLimitedMessage,
		})
		return
	}

	email := r.FormValue("email")
	_, rawToken, err := h.service.Login(r.Context(), LoginInput{
		Email:     email,
		Password:  r.FormValue("password"),
		UserAgent: r.UserAgent(),
		IPAddress: ip,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			h.renderLogin(w, http.StatusUnprocessableEntity, loginPageData{
				CSRFToken: middleware.CSRFToken(r.Context()),
				Form:      loginFormData{Email: strings.ToLower(strings.TrimSpace(email))},
				Error:     invalidCredentialsMessage,
			})
			return
		}
		h.logger.Error("login failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.loginLimit.Reset(ip)
	SetSessionCookie(w, rawToken, h.cookieSecure)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(SessionCookieName(h.cookieSecure)); err == nil {
		if err := h.service.Logout(r.Context(), c.Value); err != nil {
			h.logger.Error("logout revoke failed", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	ClearSessionCookie(w, h.cookieSecure)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) showSessions(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	currentID := ""
	if sess, ok := SessionFromContext(r.Context()); ok {
		currentID = sess.ID
	}

	sessions, err := h.service.ListSessions(r.Context(), user.ID, currentID)
	if err != nil {
		h.logger.Error("list sessions failed", "err", err, "user_id", user.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rows := make([]sessionRow, 0, len(sessions))
	for _, s := range sessions {
		rows = append(rows, sessionRow{
			ID:         s.ID,
			UserAgent:  displayOrDash(s.UserAgent),
			IPAddress:  displayOrDash(s.IPAddress),
			LastSeenAt: formatSessionTime(s.LastSeenAt),
			CreatedAt:  formatSessionTime(s.CreatedAt),
			Current:    s.Current,
		})
	}

	h.renderSessions(w, http.StatusOK, sessionsPageData{
		CSRFToken: middleware.CSRFToken(r.Context()),
		Sessions:  rows,
	})
}

func (h *Handler) revokeSession(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	sessionID := strings.TrimSpace(r.PathValue("id"))
	current, _ := SessionFromContext(r.Context())

	if err := h.service.RevokeSession(r.Context(), user.ID, sessionID); err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("revoke session failed", "err", err, "user_id", user.ID, "session_id", sessionID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if current.ID != "" && current.ID == sessionID {
		ClearSessionCookie(w, h.cookieSecure)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/account/sessions", http.StatusSeeOther)
}

func (h *Handler) showForgotPassword(w http.ResponseWriter, r *http.Request) {
	h.renderForgotPassword(w, http.StatusOK, forgotPasswordPageData{
		CSRFToken: middleware.CSRFToken(r.Context()),
	})
}

func (h *Handler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	rawToken, user, err := h.service.RequestPasswordReset(r.Context(), email)
	if err != nil {
		h.logger.Error("request password reset failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if rawToken != "" {
		resetURL := requestBaseURL(r) + "/reset-password/" + url.PathEscape(rawToken)
		if err := h.mailer.Send(mail.Message{
			To:      user.Email,
			Subject: "Reset your Forgeboard password",
			Body:    passwordResetEmailBody(user.DisplayName, resetURL),
		}); err != nil {
			h.logger.Error("send password reset email failed", "err", err, "user_id", user.ID)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// Always show the same acknowledgment — do not reveal whether the account exists.
	h.renderForgotPassword(w, http.StatusOK, forgotPasswordPageData{
		CSRFToken: middleware.CSRFToken(r.Context()),
		Form:      forgotPasswordFormData{Email: email},
		Message:   forgotPasswordAckMessage,
	})
}

func (h *Handler) showResetPassword(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.PathValue("token"))
	h.renderResetPassword(w, http.StatusOK, resetPasswordPageData{
		CSRFToken: middleware.CSRFToken(r.Context()),
		Token:     token,
	})
}

func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	token := strings.TrimSpace(r.PathValue("token"))
	fieldErrs, err := h.service.ResetPassword(r.Context(), ResetPasswordInput{
		Token:                token,
		Password:             r.FormValue("password"),
		PasswordConfirmation: r.FormValue("password_confirmation"),
	})
	if err != nil {
		if errors.Is(err, ErrInvalidResetToken) {
			h.renderResetPassword(w, http.StatusBadRequest, resetPasswordPageData{
				CSRFToken: middleware.CSRFToken(r.Context()),
				Token:     token,
				Errors:    ResetPasswordErrors{Token: "This reset link is invalid or has expired."},
			})
			return
		}
		h.logger.Error("reset password failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if fieldErrs.Any() {
		status := http.StatusUnprocessableEntity
		if fieldErrs.Token != "" {
			status = http.StatusBadRequest
		}
		h.renderResetPassword(w, status, resetPasswordPageData{
			CSRFToken: middleware.CSRFToken(r.Context()),
			Token:     token,
			Errors:    fieldErrs,
		})
		return
	}

	h.renderResetPassword(w, http.StatusOK, resetPasswordPageData{
		CSRFToken: middleware.CSRFToken(r.Context()),
		Token:     token,
		Success:   true,
	})
}

func (h *Handler) renderRegister(w http.ResponseWriter, status int, data registerPageData) {
	if err := h.render.Render(w, status, "register", data); err != nil {
		h.logger.Error("render register failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) renderLogin(w http.ResponseWriter, status int, data loginPageData) {
	if err := h.render.Render(w, status, "login", data); err != nil {
		h.logger.Error("render login failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) renderVerifyEmailSuccess(w http.ResponseWriter, status int) {
	if err := h.render.Render(w, status, "verify_email", nil); err != nil {
		h.logger.Error("render verify email success failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) renderVerifyEmailError(w http.ResponseWriter, status int) {
	if err := h.render.Render(w, status, "verify_email_error", nil); err != nil {
		h.logger.Error("render verify email error failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) renderForgotPassword(w http.ResponseWriter, status int, data forgotPasswordPageData) {
	if err := h.render.Render(w, status, "forgot_password", data); err != nil {
		h.logger.Error("render forgot password failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) renderResetPassword(w http.ResponseWriter, status int, data resetPasswordPageData) {
	if err := h.render.Render(w, status, "reset_password", data); err != nil {
		h.logger.Error("render reset password failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) renderSessions(w http.ResponseWriter, status int, data sessionsPageData) {
	if err := h.render.Render(w, status, "account_sessions", data); err != nil {
		h.logger.Error("render account sessions failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func displayOrDash(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "—"
	}
	return v
}

func formatSessionTime(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.UTC().Format("2006-01-02 15:04 UTC")
}

func acceptedTerms(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "on", "true", "1", "yes":
		return true
	default:
		return false
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

func verificationEmailBody(displayName, verifyURL string) string {
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = "there"
	}
	return fmt.Sprintf(
		"Hi %s,\n\nPlease verify your Forgeboard email by opening this link:\n\n%s\n\nThis link expires in 24 hours.\n",
		name,
		verifyURL,
	)
}

func passwordResetEmailBody(displayName, resetURL string) string {
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = "there"
	}
	return fmt.Sprintf(
		"Hi %s,\n\nReset your Forgeboard password by opening this link:\n\n%s\n\nThis link expires in 1 hour. If you did not request a reset, you can ignore this email.\n",
		name,
		resetURL,
	)
}
