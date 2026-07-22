package auth

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/render"
)

// Handler serves auth HTTP endpoints.
type Handler struct {
	service *Service
	render  *render.Renderer
	logger  *slog.Logger
}

// NewHandler constructs an auth HTTP handler.
func NewHandler(service *Service, renderer *render.Renderer, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		render:  renderer,
		logger:  logger,
	}
}

// Mount registers auth routes on mux.
func (h *Handler) Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /register", h.showRegister)
	mux.HandleFunc("POST /register", h.register)
}

type registerPageData struct {
	Form   registerFormData
	Errors RegisterErrors
}

type registerFormData struct {
	DisplayName string
	Email       string
	AcceptTerms bool
}

func (h *Handler) showRegister(w http.ResponseWriter, r *http.Request) {
	h.renderRegister(w, http.StatusOK, registerPageData{})
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
			Form: registerFormData{
				DisplayName: strings.TrimSpace(in.DisplayName),
				Email:       strings.ToLower(strings.TrimSpace(in.Email)),
				AcceptTerms: in.AcceptTerms,
			},
			Errors: fieldErrs,
		})
		return
	}

	// Session creation is deferred to STEP-12; redirect stub until then.
	h.logger.Info("user registered", "user_id", user.ID, "email", user.Email)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) renderRegister(w http.ResponseWriter, status int, data registerPageData) {
	if err := h.render.Render(w, status, "register", data); err != nil {
		h.logger.Error("render register failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func acceptedTerms(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "on", "true", "1", "yes":
		return true
	default:
		return false
	}
}
