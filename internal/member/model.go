package member

import "time"

// Membership is a user's role within a workspace.
type Membership struct {
	ID          string
	WorkspaceID string
	UserID      string
	Role        Role
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Access is a resolved workspace membership used by authorization middleware.
type Access struct {
	WorkspaceID   string
	WorkspaceName string
	WorkspaceSlug string
	Membership    Membership
}
