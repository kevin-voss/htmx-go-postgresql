package ui

// Nav section identifiers for workspace_nav active state.
const (
	NavProjects = "projects"
	NavLabels   = "labels"
	NavMembers  = "members"
	NavSettings = "settings"
)

// Crumb is one breadcrumb trail entry. Empty Href means the current page.
type Crumb struct {
	Label string
	Href  string
}

// Chrome carries shared app header / nav / breadcrumb data for authenticated pages.
type Chrome struct {
	DisplayName      string
	CSRFToken        string
	WorkspaceName    string
	WorkspaceSlug    string
	Role             string
	NavActive        string
	ShowWorkspaceNav bool
	Breadcrumbs      []Crumb
}

// App builds chrome for account-level pages (no workspace nav).
func App(displayName, csrf string, crumbs ...Crumb) Chrome {
	return Chrome{
		DisplayName: displayName,
		CSRFToken:   csrf,
		Breadcrumbs: crumbs,
	}
}

// Workspace builds chrome for workspace-scoped pages.
func Workspace(displayName, csrf, name, slug, role, navActive string, crumbs ...Crumb) Chrome {
	return Chrome{
		DisplayName:      displayName,
		CSRFToken:        csrf,
		WorkspaceName:    name,
		WorkspaceSlug:    slug,
		Role:             role,
		NavActive:        navActive,
		ShowWorkspaceNav: slug != "",
		Breadcrumbs:      crumbs,
	}
}
