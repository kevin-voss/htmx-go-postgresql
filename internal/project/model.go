package project

import "time"

// Project is a workspace-scoped container for issues.
type Project struct {
	ID          string
	WorkspaceID string
	Name        string
	Slug        string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
