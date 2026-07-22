package comment

import (
	"strings"
	"unicode/utf8"
)

const (
	bodyMin = 1
	bodyMax = 10000
)

// CreateInput is the public create-comment form payload.
type CreateInput struct {
	WorkspaceID string
	ProjectID   string
	IssueID     string
	AuthorID    string
	Body        string
}

// CreateErrors holds per-field validation messages for the create form.
type CreateErrors struct {
	Body string
}

// Any reports whether any field error is set.
func (e CreateErrors) Any() bool {
	return e.Body != ""
}

// ValidateCreate applies create-comment field rules.
func ValidateCreate(in CreateInput) CreateErrors {
	var errs CreateErrors

	bodyLen := utf8.RuneCountInString(in.Body)
	if bodyLen < bodyMin {
		errs.Body = "Comment is required."
	} else if bodyLen > bodyMax {
		errs.Body = "Comment must be at most 10000 characters."
	}

	return errs
}

func normalizeCreateInput(in CreateInput) CreateInput {
	return CreateInput{
		WorkspaceID: strings.TrimSpace(in.WorkspaceID),
		ProjectID:   strings.TrimSpace(in.ProjectID),
		IssueID:     strings.TrimSpace(in.IssueID),
		AuthorID:    strings.TrimSpace(in.AuthorID),
		Body:        strings.TrimSpace(in.Body),
	}
}
