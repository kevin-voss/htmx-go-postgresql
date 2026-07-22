package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
)

func TestCSRFIssuesCookieAndAllowsSafeMethods(t *testing.T) {
	t.Parallel()

	h := middleware.CSRF(false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := middleware.CSRFToken(r.Context()); got == "" {
			t.Fatal("expected CSRF token in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	cookie := cookieValue(rr, "forgeboard_csrf")
	if cookie == "" {
		t.Fatal("expected forgeboard_csrf cookie")
	}
}

func TestCSRFRejectsPOSTWithoutToken(t *testing.T) {
	t.Parallel()

	h := middleware.CSRF(false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatal("handler should not run for unsafe methods without CSRF")
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Seed cookie via GET first.
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, httptest.NewRequest(http.MethodGet, "/", nil))
	token := cookieValue(getRR, "forgeboard_csrf")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("x=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "forgeboard_csrf", Value: token})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestCSRFAcceptsMatchingFormToken(t *testing.T) {
	t.Parallel()

	h := middleware.CSRF(false)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, httptest.NewRequest(http.MethodGet, "/", nil))
	token := cookieValue(getRR, "forgeboard_csrf")

	form := strings.NewReader(middleware.CSRFFieldName + "=" + token)
	req := httptest.NewRequest(http.MethodPost, "/", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "forgeboard_csrf", Value: token})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
}

func TestCSRFAcceptsHeaderToken(t *testing.T) {
	t.Parallel()

	h := middleware.CSRF(false)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, httptest.NewRequest(http.MethodGet, "/", nil))
	token := cookieValue(getRR, "forgeboard_csrf")

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(middleware.CSRFHeaderName, token)
	req.AddCookie(&http.Cookie{Name: "forgeboard_csrf", Value: token})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
}

func cookieValue(rr *httptest.ResponseRecorder, name string) string {
	for _, c := range rr.Result().Cookies() {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}
