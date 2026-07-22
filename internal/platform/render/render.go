package render

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Renderer executes html/template layouts and pages from an fs.FS.
type Renderer struct {
	templates map[string]*template.Template
}

// New parses layouts, shared partials, and each page into its own template set
// so page-level title/content blocks do not overwrite each other.
func New(fsys fs.FS) (*Renderer, error) {
	layouts, err := fs.Glob(fsys, "templates/layouts/*.html")
	if err != nil {
		return nil, fmt.Errorf("render: glob layouts: %w", err)
	}
	if len(layouts) == 0 {
		return nil, fmt.Errorf("render: no layout templates found")
	}

	pages, err := fs.Glob(fsys, "templates/pages/*.html")
	if err != nil {
		return nil, fmt.Errorf("render: glob pages: %w", err)
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("render: no page templates found")
	}

	shared := append([]string{}, layouts...)
	for _, pattern := range []string{
		"templates/components/*.html",
		"templates/fragments/*.html",
	} {
		files, err := fs.Glob(fsys, pattern)
		if err != nil {
			return nil, fmt.Errorf("render: glob %s: %w", pattern, err)
		}
		shared = append(shared, files...)
	}

	templates := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		name := pageName(page)
		files := append(append([]string{}, shared...), page)
		t, err := template.New(name).ParseFS(fsys, files...)
		if err != nil {
			return nil, fmt.Errorf("render: parse %s: %w", page, err)
		}
		templates[name] = t
	}

	return &Renderer{templates: templates}, nil
}

// Render writes the named page inside the base layout with the given status.
func (r *Renderer) Render(w http.ResponseWriter, status int, name string, data any) error {
	return r.execute(w, status, name, "base", data)
}

// RenderFragment writes a named fragment template without the base layout chrome.
// page must be a registered page that includes the fragment define (all pages
// share templates/fragments/*.html).
func (r *Renderer) RenderFragment(w http.ResponseWriter, status int, page, fragment string, data any) error {
	return r.execute(w, status, page, fragment, data)
}

func (r *Renderer) execute(w http.ResponseWriter, status int, page, tmpl string, data any) error {
	t, ok := r.templates[page]
	if !ok {
		return fmt.Errorf("render: template %q not found", page)
	}

	// Buffer first so a missing/broken template does not commit a partial response.
	var buf strings.Builder
	if err := t.ExecuteTemplate(&buf, tmpl, data); err != nil {
		return fmt.Errorf("render: execute %q (%s): %w", page, tmpl, err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, err := io.WriteString(w, buf.String())
	return err
}

func pageName(file string) string {
	base := path.Base(file)
	return strings.TrimSuffix(base, path.Ext(base))
}
