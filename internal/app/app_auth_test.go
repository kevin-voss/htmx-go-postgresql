package app_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kevin-voss/htmx-go-postgresql/internal/app"
	"github.com/kevin-voss/htmx-go-postgresql/internal/config"
)

func TestAppRequiresAuthentication(t *testing.T) {
	t.Parallel()

	application := app.New(
		config.Config{Env: "test", Address: ":0"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/app", nil)
	rr := httptest.NewRecorder()
	application.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusFound)
	}
	if loc := rr.Header().Get("Location"); loc != "/login" {
		t.Fatalf("Location = %q, want /login", loc)
	}
}

func TestAppHTMXUnauthorized(t *testing.T) {
	t.Parallel()

	application := app.New(
		config.Config{Env: "test", Address: ":0"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/app", nil)
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	application.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}
