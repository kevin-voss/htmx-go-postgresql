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
	t, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("render: template %q not found", name)
	}

	// Buffer first so a missing/broken template does not commit a partial response.
	var buf strings.Builder
	if err := t.ExecuteTemplate(&buf, "base", data); err != nil {
		return fmt.Errorf("render: execute %q: %w", name, err)
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
