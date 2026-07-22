package workspace

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/kevin-voss/htmx-go-postgresql/internal/project"
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
	Onboard(ctx context.Context, name, slug, createdBy, projectName, projectSlug string) (OnboardResult, error)
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

// OnboardInput is the first-time onboarding form payload.
type OnboardInput struct {
	Name        string
	Slug        string
	ProjectName string
	CreatedBy   string
}

// OnboardErrors holds per-field validation messages for onboarding.
type OnboardErrors struct {
	Name        string
	Slug        string
	ProjectName string
}

// Any reports whether any field error is set.
func (e OnboardErrors) Any() bool {
	return e.Name != "" || e.Slug != "" || e.ProjectName != ""
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

// Onboard validates input and creates workspace, Owner membership, and first project.
func (s *Service) Onboard(ctx context.Context, in OnboardInput) (OnboardResult, OnboardErrors, error) {
	normalized := normalizeOnboardInput(in)
	fieldErrs := ValidateOnboard(normalized)
	if fieldErrs.Any() {
		return OnboardResult{}, fieldErrs, nil
	}

	projectSlug := project.SlugFromName(normalized.ProjectName)
	result, err := s.store.Onboard(
		ctx,
		normalized.Name,
		normalized.Slug,
		normalized.CreatedBy,
		normalized.ProjectName,
		projectSlug,
	)
	if err != nil {
		if errors.Is(err, ErrDuplicateSlug) {
			return OnboardResult{}, OnboardErrors{Slug: "This slug is already taken."}, nil
		}
		return OnboardResult{}, OnboardErrors{}, fmt.Errorf("workspace: onboard: %w", err)
	}
	return result, OnboardErrors{}, nil
}

// ValidateOnboard applies onboarding field rules (no uniqueness check).
func ValidateOnboard(in OnboardInput) OnboardErrors {
	var errs OnboardErrors

	createErrs := ValidateCreate(CreateInput{Name: in.Name, Slug: in.Slug})
	errs.Name = createErrs.Name
	errs.Slug = createErrs.Slug

	projectNameLen := utf8.RuneCountInString(in.ProjectName)
	if projectNameLen < nameMin || projectNameLen > nameMax {
		errs.ProjectName = "Project name must be between 2 and 50 characters."
	}

	return errs
}

func normalizeOnboardInput(in OnboardInput) OnboardInput {
	return OnboardInput{
		Name:        strings.TrimSpace(in.Name),
		Slug:        strings.ToLower(strings.TrimSpace(in.Slug)),
		ProjectName: strings.TrimSpace(in.ProjectName),
		CreatedBy:   strings.TrimSpace(in.CreatedBy),
	}
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
