package project

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	nameMin = 2
	nameMax = 50
	slugMin = 2
	slugMax = 48
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Store is the persistence port used by Service.
type Store interface {
	Create(ctx context.Context, workspaceID, name, slug, createdBy string) (Project, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]Project, error)
	GetByWorkspaceAndSlug(ctx context.Context, workspaceID, slug string) (Project, error)
	GetByID(ctx context.Context, id string) (Project, error)
}

// CreateInput is the public create-project form payload.
type CreateInput struct {
	WorkspaceID string
	Name        string
	Slug        string
	CreatedBy   string
}

// CreateErrors holds per-field validation messages for the create form.
type CreateErrors struct {
	Name string
	Slug string
}

// Any reports whether any field error is set.
func (e CreateErrors) Any() bool {
	return e.Name != "" || e.Slug != ""
}

// Service implements project business rules.
type Service struct {
	store Store
}

// NewService constructs a project service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// ListByWorkspace returns projects for a workspace ordered by name.
func (s *Service) ListByWorkspace(ctx context.Context, workspaceID string) ([]Project, error) {
	if workspaceID == "" {
		return nil, nil
	}
	return s.store.ListByWorkspace(ctx, workspaceID)
}

// GetByWorkspaceAndSlug returns a project, or ErrNotFound.
func (s *Service) GetByWorkspaceAndSlug(ctx context.Context, workspaceID, slug string) (Project, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if workspaceID == "" || slug == "" {
		return Project{}, ErrNotFound
	}
	return s.store.GetByWorkspaceAndSlug(ctx, workspaceID, slug)
}

// GetByID returns a project, or ErrNotFound.
func (s *Service) GetByID(ctx context.Context, id string) (Project, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Project{}, ErrNotFound
	}
	return s.store.GetByID(ctx, id)
}

// Create validates and persists a project under a workspace.
// On validation or duplicate-slug failure it returns field errors and a zero Project.
func (s *Service) Create(ctx context.Context, in CreateInput) (Project, CreateErrors, error) {
	normalized := normalizeCreateInput(in)
	fieldErrs := ValidateCreate(normalized)
	if fieldErrs.Any() {
		return Project{}, fieldErrs, nil
	}

	p, err := s.store.Create(ctx, normalized.WorkspaceID, normalized.Name, normalized.Slug, normalized.CreatedBy)
	if err != nil {
		if errors.Is(err, ErrDuplicateSlug) {
			return Project{}, CreateErrors{Slug: "This project slug is already taken in this workspace."}, nil
		}
		return Project{}, CreateErrors{}, fmt.Errorf("project: create: %w", err)
	}
	return p, CreateErrors{}, nil
}

// ValidateCreate applies create-project field rules (no uniqueness check).
func ValidateCreate(in CreateInput) CreateErrors {
	var errs CreateErrors

	nameLen := utf8.RuneCountInString(in.Name)
	if nameLen < nameMin || nameLen > nameMax {
		errs.Name = "Name must be between 2 and 50 characters."
	}

	slugLen := utf8.RuneCountInString(in.Slug)
	switch {
	case slugLen < slugMin || slugLen > slugMax:
		errs.Slug = "Slug must be between 2 and 48 characters."
	case !slugPattern.MatchString(in.Slug):
		errs.Slug = "Slug may only use lowercase letters, numbers, and hyphens."
	}

	return errs
}

func normalizeCreateInput(in CreateInput) CreateInput {
	return CreateInput{
		WorkspaceID: strings.TrimSpace(in.WorkspaceID),
		Name:        strings.TrimSpace(in.Name),
		Slug:        strings.ToLower(strings.TrimSpace(in.Slug)),
		CreatedBy:   strings.TrimSpace(in.CreatedBy),
	}
}

// SlugFromName derives a URL slug from a project name.
func SlugFromName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return "project"
	}

	var b strings.Builder
	prevHyphen := false
	for _, r := range name {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			prevHyphen = false
		default:
			if b.Len() > 0 && !prevHyphen {
				b.WriteByte('-')
				prevHyphen = true
			}
		}
	}

	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "project"
	}
	if utf8.RuneCountInString(slug) > slugMax {
		slug = string([]rune(slug)[:slugMax])
		slug = strings.Trim(slug, "-")
	}
	if utf8.RuneCountInString(slug) < slugMin {
		return "project"
	}
	return slug
}
