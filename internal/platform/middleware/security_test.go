package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
)

func TestSecurityHeaders(t *testing.T) {
	t.Parallel()

	h := middleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	checks := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"X-XSS-Protection":       "0",
	}
	for header, want := range checks {
		if got := rr.Header().Get(header); got != want {
			t.Fatalf("%s = %q, want %q", header, got, want)
		}
	}
}
