package integration

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/kevin-voss/htmx-go-postgresql/internal/app"
	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/config"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
)

func TestLoginHandlerIssuesSession(t *testing.T) {
	pool := Pool(t)
	fx := SeedWorkspace(t, pool)

	application := app.New(
		config.Config{Env: "test", Address: ":0", CookieSecure: false},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		pool,
	)

	csrfCookie, csrfToken := fetchCSRF(t, application, "/login")

	form := url.Values{}
	form.Set("email", fx.Email)
	form.Set("password", fx.Password)
	form.Set(middleware.CSRFFieldName, csrfToken)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(csrfCookie)
	rr := httptest.NewRecorder()
	application.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body=%q", rr.Code, http.StatusSeeOther, rr.Body.String())
	}

	sessionCookie := findCookie(rr, auth.SessionCookieName(false))
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatal("missing session cookie after login")
	}

	appReq := httptest.NewRequest(http.MethodGet, "/app", nil)
	appReq.AddCookie(sessionCookie)
	appRR := httptest.NewRecorder()
	application.Routes().ServeHTTP(appRR, appReq)
	if appRR.Code != http.StatusOK && appRR.Code != http.StatusSeeOther {
		t.Fatalf("authenticated /app status = %d, want 200 or 303; body=%q", appRR.Code, appRR.Body.String())
	}
}

func TestCreateIssueHTMXPartial(t *testing.T) {
	pool := Pool(t)
	fx := SeedWorkspace(t, pool)

	application := app.New(
		config.Config{Env: "test", Address: ":0", CookieSecure: false},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		pool,
	)

	sessionCookie := loginSession(t, application, fx.Email, fx.Password)
	csrfCookie, csrfToken := fetchCSRFAuthed(t, application, sessionCookie, "/w/"+fx.WorkspaceSlug+"/projects/"+fx.ProjectSlug+"/issues")

	title := "HTMX integration issue " + uniqueSuffix(t)
	form := url.Values{}
	form.Set("title", title)
	form.Set("description", "created via httptest")
	form.Set(middleware.CSRFFieldName, csrfToken)

	path := "/w/" + fx.WorkspaceSlug + "/projects/" + fx.ProjectSlug + "/issues"
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Request-Type", "partial")
	req.AddCookie(sessionCookie)
	req.AddCookie(csrfCookie)
	rr := httptest.NewRecorder()
	application.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%q", rr.Code, http.StatusCreated, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, title) {
		t.Fatalf("HTMX fragment missing title %q; body=%q", title, body)
	}
}

func loginSession(t *testing.T, application *app.Application, email, password string) *http.Cookie {
	t.Helper()

	csrfCookie, csrfToken := fetchCSRF(t, application, "/login")
	form := url.Values{}
	form.Set("email", email)
	form.Set("password", password)
	form.Set(middleware.CSRFFieldName, csrfToken)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(csrfCookie)
	rr := httptest.NewRecorder()
	application.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("login status = %d; body=%q", rr.Code, rr.Body.String())
	}
	session := findCookie(rr, auth.SessionCookieName(false))
	if session == nil {
		t.Fatal("login missing session cookie")
	}
	return session
}

func fetchCSRF(t *testing.T, application *app.Application, path string) (*http.Cookie, string) {
	t.Helper()

	rr := httptest.NewRecorder()
	application.Routes().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
	if rr.Code != http.StatusOK && rr.Code != http.StatusFound && rr.Code != http.StatusSeeOther {
		t.Fatalf("csrf seed status = %d for %s", rr.Code, path)
	}
	cookie := findCookie(rr, "forgeboard_csrf")
	if cookie == nil || cookie.Value == "" {
		t.Fatal("missing forgeboard_csrf cookie")
	}
	return cookie, cookie.Value
}

func fetchCSRFAuthed(t *testing.T, application *app.Application, session *http.Cookie, path string) (*http.Cookie, string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.AddCookie(session)
	rr := httptest.NewRecorder()
	application.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("authed csrf seed status = %d for %s; body=%q", rr.Code, path, rr.Body.String())
	}
	cookie := findCookie(rr, "forgeboard_csrf")
	if cookie == nil || cookie.Value == "" {
		t.Fatal("missing forgeboard_csrf cookie")
	}
	return cookie, cookie.Value
}

func findCookie(rr *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range rr.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}
