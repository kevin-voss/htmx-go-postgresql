package comment

import "time"

// Comment is an issue-scoped discussion message.
type Comment struct {
	ID         string
	IssueID    string
	AuthorID   string
	AuthorName string
	Body       string
	CreatedAt  time.Time
}

// CanDelete reports whether actorID with workspaceRole may delete c.
// Authors may delete their own comments; Owners (elevated) may delete any.
func (c Comment) CanDelete(actorID, workspaceRole string) bool {
	if actorID != "" && c.AuthorID == actorID {
		return true
	}
	return workspaceRole == "owner"
}
