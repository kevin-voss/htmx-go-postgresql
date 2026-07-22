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

func TestRenderProjectShowFullAndFragment(t *testing.T) {
	t.Parallel()

	r, err := render.New(web.Templates)
	if err != nil {
		t.Fatalf("render.New: %v", err)
	}

	data := map[string]any{
		"WorkspaceName": "Acme",
		"WorkspaceSlug": "acme",
		"Project": map[string]any{
			"Name": "Platform",
			"Slug": "platform",
		},
		"User": map[string]any{
			"DisplayName": "Ada",
		},
		"Role": "owner",
	}

	full := httptest.NewRecorder()
	if err := r.Render(full, http.StatusOK, "project_show", data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	fullBody := full.Body.String()
	if !strings.Contains(fullBody, "<!DOCTYPE html>") {
		t.Fatalf("full page missing layout chrome, got %q", fullBody)
	}
	if !strings.Contains(fullBody, `src="/static/vendor/htmx-4.0.0-beta5.min.js"`) {
		t.Fatalf("full page missing local HTMX script, got %q", fullBody)
	}
	if !strings.Contains(fullBody, `id="project-content"`) {
		t.Fatalf("full page missing project content, got %q", fullBody)
	}

	frag := httptest.NewRecorder()
	if err := r.RenderFragment(frag, http.StatusOK, "project_show", "project_content", data); err != nil {
		t.Fatalf("RenderFragment: %v", err)
	}
	fragBody := frag.Body.String()
	if strings.Contains(fragBody, "<!DOCTYPE html>") || strings.Contains(fragBody, "<html") {
		t.Fatalf("fragment unexpectedly includes layout chrome, got %q", fragBody)
	}
	if !strings.Contains(fragBody, `id="project-content"`) {
		t.Fatalf("fragment missing project content, got %q", fragBody)
	}
	if !strings.Contains(fragBody, "Platform") {
		t.Fatalf("fragment missing project name, got %q", fragBody)
	}
}
