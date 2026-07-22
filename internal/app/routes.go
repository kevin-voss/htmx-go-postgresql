package app

import (
	"io/fs"
	"net/http"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/member"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/ui"
	"github.com/kevin-voss/htmx-go-postgresql/web"
)

// Routes returns the root handler with middleware chains applied.
func (a *Application) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", a.home)
	mux.HandleFunc("GET /health", a.health)
	a.Auth.Mount(mux)
	a.Workspace.Mount(mux)
	a.Project.Mount(mux)
	a.Issue.Mount(mux)
	a.Comment.Mount(mux)
	a.MemberHTTP.Mount(mux)

	staticRoot, err := fs.Sub(web.Static, "static")
	if err != nil {
		// embed layout is fixed at build time; fail fast if it drifts.
		panic("static embed root missing: " + err.Error())
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticRoot)))

	mux.Handle("GET /app", auth.RequireAuthentication(http.HandlerFunc(a.appDashboard)))

	return middleware.Chain(
		mux,
		middleware.SecurityHeaders,
		a.Auth.LoadSessionMiddleware(),
		middleware.CSRF(a.Config.CookieSecure),
	)
}

func (a *Application) home(w http.ResponseWriter, r *http.Request) {
	if err := a.Render.Render(w, http.StatusOK, "home", nil); err != nil {
		a.Logger.Error("render home failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (a *Application) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

type appPageData struct {
	CSRFToken  string
	User       auth.User
	Workspaces []member.UserWorkspace
	Chrome     ui.Chrome
}

func (a *Application) appDashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	workspaces, err := a.Members.ListWorkspacesForUser(r.Context(), user.ID)
	if err != nil {
		a.Logger.Error("list workspaces failed", "err", err, "user_id", user.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if len(workspaces) == 0 {
		http.Redirect(w, r, "/app/onboarding", http.StatusSeeOther)
		return
	}

	csrf := middleware.CSRFToken(r.Context())
	data := appPageData{
		CSRFToken:  csrf,
		User:       user,
		Workspaces: workspaces,
		Chrome:     ui.App(user.DisplayName, csrf, ui.Crumb{Label: "App"}),
	}
	if err := a.Render.Render(w, http.StatusOK, "app", data); err != nil {
		a.Logger.Error("render app dashboard failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
