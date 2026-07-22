package member

import (
	"context"
	"errors"
	"testing"
)

type stubStore struct {
	access map[string]Access // key: slug|userID
	create func(ctx context.Context, workspaceID, userID string, role Role) (Membership, error)
}

func (s *stubStore) Create(ctx context.Context, workspaceID, userID string, role Role) (Membership, error) {
	if s.create != nil {
		return s.create(ctx, workspaceID, userID, role)
	}
	return Membership{ID: "m1", WorkspaceID: workspaceID, UserID: userID, Role: role}, nil
}

func (s *stubStore) GetByWorkspaceAndUser(context.Context, string, string) (Membership, error) {
	return Membership{}, ErrNotFound
}

func (s *stubStore) GetAccessBySlug(_ context.Context, slug, userID string) (Access, error) {
	if s.access == nil {
		return Access{}, ErrNotFound
	}
	a, ok := s.access[slug+"|"+userID]
	if !ok {
		return Access{}, ErrNotFound
	}
	return a, nil
}

func TestResolveAccessBySlugNotFound(t *testing.T) {
	t.Parallel()

	svc := NewService(&stubStore{})
	_, err := svc.ResolveAccessBySlug(context.Background(), "acme", "u1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestResolveAccessBySlugOK(t *testing.T) {
	t.Parallel()

	store := &stubStore{
		access: map[string]Access{
			"acme|u1": {
				WorkspaceID:   "w1",
				WorkspaceName: "Acme",
				WorkspaceSlug: "acme",
				Membership:    Membership{ID: "m1", WorkspaceID: "w1", UserID: "u1", Role: RoleMember},
			},
		},
	}
	svc := NewService(store)

	access, err := svc.ResolveAccessBySlug(context.Background(), "  Acme  ", "u1")
	if err != nil {
		t.Fatalf("ResolveAccessBySlug: %v", err)
	}
	if access.WorkspaceSlug != "acme" || access.Membership.Role != RoleMember {
		t.Fatalf("unexpected access: %+v", access)
	}
}

func TestAddMemberRejectsInvalidRole(t *testing.T) {
	t.Parallel()

	svc := NewService(&stubStore{})
	_, err := svc.AddMember(context.Background(), "w1", "u1", Role("admin"))
	if err == nil {
		t.Fatal("want error for invalid role")
	}
}
