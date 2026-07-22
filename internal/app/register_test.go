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

func TestRegisterGET(t *testing.T) {
	t.Parallel()

	application := app.New(
		config.Config{Env: "test", Address: ":0"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rr := httptest.NewRecorder()
	application.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `name="display_name"`) {
		t.Fatalf("missing display_name field, body=%q", body)
	}
	if !strings.Contains(body, `name="email"`) {
		t.Fatalf("missing email field, body=%q", body)
	}
	if !strings.Contains(body, `name="password"`) {
		t.Fatalf("missing password field, body=%q", body)
	}
	if !strings.Contains(body, `name="terms"`) {
		t.Fatalf("missing terms checkbox, body=%q", body)
	}
}

func TestRegisterPOSTInvalidShowsFieldErrors(t *testing.T) {
	t.Parallel()

	application := app.New(
		config.Config{Env: "test", Address: ":0"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil,
	)

	form := strings.NewReader(strings.Join([]string{
		"display_name=A",
		"email=bad",
		"password=short",
		"password_confirmation=other",
	}, "&"))
	req := httptest.NewRequest(http.MethodPost, "/register", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	application.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d; body=%q", rr.Code, http.StatusUnprocessableEntity, rr.Body.String())
	}
	body := rr.Body.String()
	for _, want := range []string{
		"Display name must be between 2 and 50 characters.",
		"Enter a valid email address.",
		"Password must be at least 12 characters.",
		"Passwords do not match.",
		"You must accept the terms.",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q; got %q", want, body)
		}
	}
}
