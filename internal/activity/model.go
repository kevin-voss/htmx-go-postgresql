package activity

import "time"

// Event types recorded for collaboration history.
const (
	TypeIssueCreated       = "issue.created"
	TypeIssueStatusChanged = "issue.status_changed"
	TypeCommentCreated     = "comment.created"
)

// Event is a workspace/project-scoped activity row.
type Event struct {
	ID          string
	WorkspaceID string
	ProjectID   string
	IssueID     string
	ActorID     string
	ActorName   string
	Type        string
	Summary     string
	CreatedAt   time.Time
}

// EventInput is the payload for recording a new activity event.
type EventInput struct {
	WorkspaceID string
	ProjectID   string
	IssueID     string
	ActorID     string
	Type        string
	Summary     string
}

// TypeLabel returns a short human-readable label for an event type.
func TypeLabel(eventType string) string {
	switch eventType {
	case TypeIssueCreated:
		return "Issue created"
	case TypeIssueStatusChanged:
		return "Status changed"
	case TypeCommentCreated:
		return "Comment added"
	default:
		return eventType
	}
}
