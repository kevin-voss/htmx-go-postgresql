package app_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kevin-voss/htmx-go-postgresql/internal/app"
	"github.com/kevin-voss/htmx-go-postgresql/internal/config"
)

func TestHome(t *testing.T) {
	t.Parallel()

	application := app.New(
		config.Config{Env: "test", Address: ":0"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	application.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(strings.ToLower(body), "forgeboard") {
		t.Fatalf("body missing Forgeboard, got %q", body)
	}
	if !strings.Contains(body, `href="/register"`) {
		t.Fatalf("body missing /register CTA, got %q", body)
	}
	if !strings.Contains(body, `href="/login"`) {
		t.Fatalf("body missing /login CTA, got %q", body)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}
}
