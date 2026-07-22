package issue

import (
	"strings"
	"time"
	"unicode/utf8"
)

const (
	labelNameMin  = 1
	labelNameMax  = 40
	defaultColor  = "#64748b"
)

// Label is a workspace-scoped tag that can be attached to issues.
type Label struct {
	ID          string
	WorkspaceID string
	Name        string
	Color       string
	CreatedAt   time.Time
}

// CreateLabelInput is the public create-label form payload.
type CreateLabelInput struct {
	WorkspaceID string
	Name        string
	Color       string
}

// CreateLabelErrors holds per-field validation messages for create label.
type CreateLabelErrors struct {
	Name  string
	Color string
}

// Any reports whether any field error is set.
func (e CreateLabelErrors) Any() bool {
	return e.Name != "" || e.Color != ""
}

// ValidateCreateLabel applies create-label field rules.
func ValidateCreateLabel(in CreateLabelInput) CreateLabelErrors {
	var errs CreateLabelErrors

	nameLen := utf8.RuneCountInString(in.Name)
	if nameLen < labelNameMin {
		errs.Name = "Name is required."
	} else if nameLen > labelNameMax {
		errs.Name = "Name must be at most 40 characters."
	}

	if in.Color != "" && !ValidLabelColor(in.Color) {
		errs.Color = "Color must be a hex value like #64748b."
	}

	return errs
}

// ValidLabelColor reports whether color is a 6-digit hex color (#RRGGBB).
func ValidLabelColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		c := color[i]
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}

func normalizeCreateLabelInput(in CreateLabelInput) CreateLabelInput {
	color := strings.TrimSpace(in.Color)
	if color == "" {
		color = defaultColor
	}
	return CreateLabelInput{
		WorkspaceID: strings.TrimSpace(in.WorkspaceID),
		Name:        strings.TrimSpace(in.Name),
		Color:       color,
	}
}
