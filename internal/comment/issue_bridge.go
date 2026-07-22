package comment

import (
	"context"

	"github.com/kevin-voss/htmx-go-postgresql/internal/issue"
)

// IssueShowLister adapts Service to issue.CommentLister.
type IssueShowLister struct {
	Service *Service
}

// ListByIssue returns comments shaped for the issue show page.
func (l IssueShowLister) ListByIssue(ctx context.Context, issueID string) ([]issue.ShowComment, error) {
	comments, err := l.Service.ListByIssue(ctx, issueID)
	if err != nil {
		return nil, err
	}
	out := make([]issue.ShowComment, 0, len(comments))
	for _, c := range comments {
		out = append(out, issue.ShowComment{
			ID:         c.ID,
			AuthorID:   c.AuthorID,
			AuthorName: c.AuthorName,
			Body:       c.Body,
			CreatedAt:  c.CreatedAt,
		})
	}
	return out, nil
}

// CountByIssue returns the comment count for an issue.
func (l IssueShowLister) CountByIssue(ctx context.Context, issueID string) (int, error) {
	return l.Service.CountByIssue(ctx, issueID)
}
