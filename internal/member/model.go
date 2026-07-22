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

// MemberView is a membership row enriched for listing UI.
type MemberView struct {
	Membership
	Email       string
	DisplayName string
}

// Access is a resolved workspace membership used by authorization middleware.
type Access struct {
	WorkspaceID   string
	WorkspaceName string
	WorkspaceSlug string
	Membership    Membership
}

// Invitation is a pending or accepted workspace invitation.
type Invitation struct {
	ID            string
	WorkspaceID   string
	Email         string
	Role          Role
	InvitedBy     string
	TokenHash     string
	CreatedAt     time.Time
	ExpiresAt     time.Time
	AcceptedAt    *time.Time
	WorkspaceName string
	WorkspaceSlug string
}
