package render_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/render"
	"github.com/kevin-voss/htmx-go-postgresql/web"
)

func TestRenderSmokePage(t *testing.T) {
	t.Parallel()

	r, err := render.New(web.Templates)
	if err != nil {
		t.Fatalf("render.New: %v", err)
	}

	rr := httptest.NewRecorder()
	if err := r.Render(rr, http.StatusOK, "smoke", nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "<title>Smoke</title>") {
		t.Fatalf("body missing title, got %q", body)
	}
	if !strings.Contains(body, "smoke ok") {
		t.Fatalf("body missing content, got %q", body)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}
}

func TestRenderUnknownTemplate(t *testing.T) {
	t.Parallel()

	r, err := render.New(web.Templates)
	if err != nil {
		t.Fatalf("render.New: %v", err)
	}

	rr := httptest.NewRecorder()
	err = r.Render(rr, http.StatusOK, "missing", nil)
	if err == nil {
		t.Fatal("Render missing template: want error")
	}
}
