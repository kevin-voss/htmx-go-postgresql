package project

import (
	"context"
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
	GetByWorkspaceAndSlug(ctx context.Context, workspaceID, slug string) (Project, error)
}

// Service implements project business rules.
type Service struct {
	store Store
}

// NewService constructs a project service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// GetByWorkspaceAndSlug returns a project, or ErrNotFound.
func (s *Service) GetByWorkspaceAndSlug(ctx context.Context, workspaceID, slug string) (Project, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if workspaceID == "" || slug == "" {
		return Project{}, ErrNotFound
	}
	return s.store.GetByWorkspaceAndSlug(ctx, workspaceID, slug)
}

// Create validates and persists a project under a workspace.
func (s *Service) Create(ctx context.Context, workspaceID, name, slug, createdBy string) (Project, error) {
	name = strings.TrimSpace(name)
	slug = strings.ToLower(strings.TrimSpace(slug))
	if err := validateName(name); err != nil {
		return Project{}, err
	}
	if err := validateSlug(slug); err != nil {
		return Project{}, err
	}
	p, err := s.store.Create(ctx, workspaceID, name, slug, createdBy)
	if err != nil {
		return Project{}, err
	}
	return p, nil
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

func validateName(name string) error {
	n := utf8.RuneCountInString(name)
	if n < nameMin || n > nameMax {
		return fmt.Errorf("project: name must be between %d and %d characters", nameMin, nameMax)
	}
	return nil
}

func validateSlug(slug string) error {
	n := utf8.RuneCountInString(slug)
	if n < slugMin || n > slugMax {
		return fmt.Errorf("project: slug must be between %d and %d characters", slugMin, slugMax)
	}
	if !slugPattern.MatchString(slug) {
		return fmt.Errorf("project: invalid slug format")
	}
	return nil
}
