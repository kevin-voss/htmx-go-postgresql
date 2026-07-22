package member

import (
	"context"
	"fmt"
	"strings"
)

// Store is the persistence port used by Service.
type Store interface {
	Create(ctx context.Context, workspaceID, userID string, role Role) (Membership, error)
	GetByWorkspaceAndUser(ctx context.Context, workspaceID, userID string) (Membership, error)
	GetAccessBySlug(ctx context.Context, slug, userID string) (Access, error)
	HasAny(ctx context.Context, userID string) (bool, error)
}

// Service implements membership lookups and role helpers.
type Service struct {
	store Store
}

// NewService constructs a membership service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// ResolveAccessBySlug returns workspace access for a member, or ErrNotFound.
func (s *Service) ResolveAccessBySlug(ctx context.Context, slug, userID string) (Access, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	userID = strings.TrimSpace(userID)
	if slug == "" || userID == "" {
		return Access{}, ErrNotFound
	}
	access, err := s.store.GetAccessBySlug(ctx, slug, userID)
	if err != nil {
		return Access{}, err
	}
	return access, nil
}

// AddMember creates a membership with the given role.
func (s *Service) AddMember(ctx context.Context, workspaceID, userID string, role Role) (Membership, error) {
	if !role.Valid() {
		return Membership{}, fmt.Errorf("member: invalid role %q", role)
	}
	m, err := s.store.Create(ctx, workspaceID, userID, role)
	if err != nil {
		return Membership{}, err
	}
	return m, nil
}

// HasAnyMembership reports whether the user belongs to at least one workspace.
func (s *Service) HasAnyMembership(ctx context.Context, userID string) (bool, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return false, nil
	}
	ok, err := s.store.HasAny(ctx, userID)
	if err != nil {
		return false, err
	}
	return ok, nil
}
