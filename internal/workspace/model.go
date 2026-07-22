package workspace

import "time"

// Workspace is a persisted tenancy container.
type Workspace struct {
	ID        string
	Name      string
	Slug      string
	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}
