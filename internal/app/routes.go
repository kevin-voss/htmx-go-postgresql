package app

import (
	"io/fs"
	"net/http"

	"github.com/kevin-voss/htmx-go-postgresql/web"
)

// Routes returns the root ServeMux with method+path patterns registered.
func (a *Application) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", a.health)

	staticRoot, err := fs.Sub(web.Static, "static")
	if err != nil {
		// embed layout is fixed at build time; fail fast if it drifts.
		panic("static embed root missing: " + err.Error())
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticRoot)))

	return mux
}

func (a *Application) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
