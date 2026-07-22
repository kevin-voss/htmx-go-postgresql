package workspace

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	nameMin = 2
	nameMax = 50
	slugMin = 2
	slugMax = 48
)

// slugPattern matches lowercase alphanumeric segments separated by single hyphens.
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Store is the persistence port used by Service.
type Store interface {
	Create(ctx context.Context, name, slug, createdBy string) (Workspace, error)
	GetBySlug(ctx context.Context, slug string) (Workspace, error)
}

// CreateInput is the public create-workspace form payload.
type CreateInput struct {
	Name      string
	Slug      string
	CreatedBy string
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

// Service implements workspace business rules.
type Service struct {
	store Store
}

// NewService constructs a workspace service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// Create validates input and persists a new workspace.
// On validation or duplicate-slug failure it returns field errors and a zero Workspace.
func (s *Service) Create(ctx context.Context, in CreateInput) (Workspace, CreateErrors, error) {
	normalized := normalizeCreateInput(in)
	fieldErrs := ValidateCreate(normalized)
	if fieldErrs.Any() {
		return Workspace{}, fieldErrs, nil
	}

	ws, err := s.store.Create(ctx, normalized.Name, normalized.Slug, normalized.CreatedBy)
	if err != nil {
		if errors.Is(err, ErrDuplicateSlug) {
			return Workspace{}, CreateErrors{Slug: "This slug is already taken."}, nil
		}
		return Workspace{}, CreateErrors{}, fmt.Errorf("workspace: create: %w", err)
	}
	return ws, CreateErrors{}, nil
}

// GetBySlug returns a workspace by slug, or ErrNotFound.
func (s *Service) GetBySlug(ctx context.Context, slug string) (Workspace, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if slug == "" {
		return Workspace{}, ErrNotFound
	}
	ws, err := s.store.GetBySlug(ctx, slug)
	if err != nil {
		return Workspace{}, err
	}
	return ws, nil
}

// ValidateCreate applies create-workspace field rules (no uniqueness check).
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
		Name:      strings.TrimSpace(in.Name),
		Slug:      strings.ToLower(strings.TrimSpace(in.Slug)),
		CreatedBy: strings.TrimSpace(in.CreatedBy),
	}
}
