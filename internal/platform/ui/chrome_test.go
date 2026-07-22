package ui_test

import (
	"testing"

	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/ui"
)

func TestAppChrome(t *testing.T) {
	t.Parallel()
	c := ui.App("Ada", "tok", ui.Crumb{Label: "App", Href: "/app"}, ui.Crumb{Label: "Sessions"})
	if c.ShowWorkspaceNav {
		t.Fatal("app chrome should not show workspace nav")
	}
	if c.CSRFToken != "tok" || c.DisplayName != "Ada" {
		t.Fatalf("unexpected chrome: %+v", c)
	}
	if len(c.Breadcrumbs) != 2 || c.Breadcrumbs[1].Href != "" {
		t.Fatalf("unexpected crumbs: %+v", c.Breadcrumbs)
	}
}

func TestWorkspaceChrome(t *testing.T) {
	t.Parallel()
	c := ui.Workspace("Ada", "tok", "Acme", "acme", "owner", ui.NavProjects,
		ui.Crumb{Label: "App", Href: "/app"},
		ui.Crumb{Label: "Acme", Href: "/w/acme"},
		ui.Crumb{Label: "Projects"},
	)
	if !c.ShowWorkspaceNav || c.NavActive != ui.NavProjects || c.WorkspaceSlug != "acme" {
		t.Fatalf("unexpected chrome: %+v", c)
	}
	if c.Role != "owner" || len(c.Breadcrumbs) != 3 {
		t.Fatalf("unexpected chrome: %+v", c)
	}
}
