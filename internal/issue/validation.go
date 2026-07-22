package issue

import (
	"strings"
	"unicode/utf8"
)

const (
	titleMin = 1
	titleMax = 200
)

// CreateInput is the public create-issue form payload.
type CreateInput struct {
	WorkspaceID string
	ProjectID   string
	Title       string
	Description string
	CreatedBy   string
}

// CreateErrors holds per-field validation messages for the create form.
type CreateErrors struct {
	Title string
}

// Any reports whether any field error is set.
func (e CreateErrors) Any() bool {
	return e.Title != ""
}

// ValidateCreate applies create-issue field rules.
func ValidateCreate(in CreateInput) CreateErrors {
	var errs CreateErrors

	titleLen := utf8.RuneCountInString(in.Title)
	if titleLen < titleMin {
		errs.Title = "Title is required."
	} else if titleLen > titleMax {
		errs.Title = "Title must be at most 200 characters."
	}

	return errs
}

func normalizeCreateInput(in CreateInput) CreateInput {
	return CreateInput{
		WorkspaceID: strings.TrimSpace(in.WorkspaceID),
		ProjectID:   strings.TrimSpace(in.ProjectID),
		Title:       strings.TrimSpace(in.Title),
		Description: strings.TrimSpace(in.Description),
		CreatedBy:   strings.TrimSpace(in.CreatedBy),
	}
}
