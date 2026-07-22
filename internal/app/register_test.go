package app_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/kevin-voss/htmx-go-postgresql/internal/app"
	"github.com/kevin-voss/htmx-go-postgresql/internal/config"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
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
	if !strings.Contains(body, `name="csrf_token"`) {
		t.Fatalf("missing csrf_token field, body=%q", body)
	}
}

func TestRegisterPOSTInvalidShowsFieldErrors(t *testing.T) {
	t.Parallel()

	application := app.New(
		config.Config{Env: "test", Address: ":0"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil,
	)

	csrfCookie, csrfToken := fetchCSRF(t, application)

	form := url.Values{}
	form.Set("display_name", "A")
	form.Set("email", "bad")
	form.Set("password", "short")
	form.Set("password_confirmation", "other")
	form.Set(middleware.CSRFFieldName, csrfToken)

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(csrfCookie)
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

func TestRegisterPOSTWithoutCSRFRejected(t *testing.T) {
	t.Parallel()

	application := app.New(
		config.Config{Env: "test", Address: ":0"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil,
	)

	form := url.Values{}
	form.Set("display_name", "Ada")
	form.Set("email", "ada@example.com")
	form.Set("password", "correct-horse-battery")
	form.Set("password_confirmation", "correct-horse-battery")
	form.Set("terms", "on")

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	application.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func fetchCSRF(t *testing.T, application *app.Application) (*http.Cookie, string) {
	t.Helper()

	rr := httptest.NewRecorder()
	application.Routes().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/register", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("csrf seed status = %d", rr.Code)
	}

	var cookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "forgeboard_csrf" {
			cookie = c
			break
		}
	}
	if cookie == nil || cookie.Value == "" {
		t.Fatal("missing forgeboard_csrf cookie")
	}
	return cookie, cookie.Value
}
