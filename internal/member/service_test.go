package member

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubStore struct {
	access       map[string]Access // key: slug|userID
	members      map[string]Membership
	invitations  map[string]Invitation
	create       func(ctx context.Context, workspaceID, userID string, role Role) (Membership, error)
	list         []MemberView
	acceptCalled bool
}

func (s *stubStore) Create(ctx context.Context, workspaceID, userID string, role Role) (Membership, error) {
	if s.create != nil {
		return s.create(ctx, workspaceID, userID, role)
	}
	return Membership{ID: "m1", WorkspaceID: workspaceID, UserID: userID, Role: role}, nil
}

func (s *stubStore) GetByWorkspaceAndUser(_ context.Context, workspaceID, userID string) (Membership, error) {
	if s.members == nil {
		return Membership{}, ErrNotFound
	}
	m, ok := s.members[workspaceID+"|"+userID]
	if !ok {
		return Membership{}, ErrNotFound
	}
	return m, nil
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

func (s *stubStore) HasAny(_ context.Context, userID string) (bool, error) {
	if s.access == nil {
		return false, nil
	}
	for key := range s.access {
		if len(key) > len(userID)+1 && key[len(key)-len(userID):] == userID {
			return true, nil
		}
	}
	return false, nil
}

func (s *stubStore) ListByWorkspace(context.Context, string) ([]MemberView, error) {
	return s.list, nil
}

func (s *stubStore) UpdateRole(_ context.Context, workspaceID, userID string, role Role) error {
	key := workspaceID + "|" + userID
	m, ok := s.members[key]
	if !ok {
		return ErrNotFound
	}
	m.Role = role
	s.members[key] = m
	return nil
}

func (s *stubStore) Delete(_ context.Context, workspaceID, userID string) error {
	key := workspaceID + "|" + userID
	if _, ok := s.members[key]; !ok {
		return ErrNotFound
	}
	delete(s.members, key)
	return nil
}

func (s *stubStore) CreateInvitation(_ context.Context, workspaceID, email string, role Role, invitedBy, tokenHash string, expiresAt time.Time) (Invitation, error) {
	inv := Invitation{
		ID:          "inv1",
		WorkspaceID: workspaceID,
		Email:       email,
		Role:        role,
		InvitedBy:   invitedBy,
		TokenHash:   tokenHash,
		ExpiresAt:   expiresAt,
	}
	if s.invitations == nil {
		s.invitations = map[string]Invitation{}
	}
	s.invitations[tokenHash] = inv
	return inv, nil
}

func (s *stubStore) GetInvitationByTokenHash(_ context.Context, tokenHash string) (Invitation, error) {
	inv, ok := s.invitations[tokenHash]
	if !ok {
		return Invitation{}, ErrNotFound
	}
	return inv, nil
}

func (s *stubStore) AcceptInvitation(_ context.Context, invitationID, workspaceID, userID string, role Role, _ time.Time) (Membership, error) {
	s.acceptCalled = true
	m := Membership{ID: "m-new", WorkspaceID: workspaceID, UserID: userID, Role: role}
	if s.members == nil {
		s.members = map[string]Membership{}
	}
	s.members[workspaceID+"|"+userID] = m
	for hash, inv := range s.invitations {
		if inv.ID == invitationID {
			now := time.Now().UTC()
			inv.AcceptedAt = &now
			s.invitations[hash] = inv
		}
	}
	return m, nil
}

func (s *stubStore) MarkInvitationAccepted(_ context.Context, invitationID string, at time.Time) error {
	for hash, inv := range s.invitations {
		if inv.ID == invitationID {
			inv.AcceptedAt = &at
			s.invitations[hash] = inv
		}
	}
	return nil
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

func TestCreateInvitationDefaultsToMember(t *testing.T) {
	t.Parallel()

	store := &stubStore{}
	svc := NewService(store)
	raw, errs, err := svc.CreateInvitation(context.Background(), CreateInvitationInput{
		WorkspaceID: "w1",
		Email:       "new@example.com",
		InvitedBy:   "owner1",
	})
	if err != nil {
		t.Fatalf("CreateInvitation: %v", err)
	}
	if errs.Any() {
		t.Fatalf("unexpected field errors: %+v", errs)
	}
	if raw == "" {
		t.Fatal("expected raw token")
	}
	inv := store.invitations[hashInvitationToken(raw)]
	if inv.Role != RoleMember {
		t.Fatalf("role = %q, want member", inv.Role)
	}
}

func TestCreateInvitationRejectsInvalidEmail(t *testing.T) {
	t.Parallel()

	svc := NewService(&stubStore{})
	_, errs, err := svc.CreateInvitation(context.Background(), CreateInvitationInput{
		WorkspaceID: "w1",
		Email:       "not-an-email",
		InvitedBy:   "owner1",
	})
	if err != nil {
		t.Fatalf("CreateInvitation: %v", err)
	}
	if errs.Email == "" {
		t.Fatal("want email field error")
	}
}

func TestAcceptInvitationCreatesMembership(t *testing.T) {
	t.Parallel()

	store := &stubStore{invitations: map[string]Invitation{}}
	raw := "raw-token-value"
	store.invitations[hashInvitationToken(raw)] = Invitation{
		ID:            "inv1",
		WorkspaceID:   "w1",
		Email:         "invitee@example.com",
		Role:          RoleMember,
		ExpiresAt:     time.Now().UTC().Add(time.Hour),
		WorkspaceSlug: "acme",
		WorkspaceName: "Acme",
	}
	svc := NewService(store)

	result, err := svc.AcceptInvitation(context.Background(), raw, "u2", "invitee@example.com")
	if err != nil {
		t.Fatalf("AcceptInvitation: %v", err)
	}
	if result.Already || result.Membership.UserID != "u2" || result.Membership.Role != RoleMember {
		t.Fatalf("unexpected result: %+v", result)
	}
	if !store.acceptCalled {
		t.Fatal("expected AcceptInvitation store call")
	}
}

func TestAcceptInvitationEmailMismatch(t *testing.T) {
	t.Parallel()

	store := &stubStore{invitations: map[string]Invitation{}}
	raw := "raw-token-value"
	store.invitations[hashInvitationToken(raw)] = Invitation{
		ID:          "inv1",
		WorkspaceID: "w1",
		Email:       "invitee@example.com",
		Role:        RoleMember,
		ExpiresAt:   time.Now().UTC().Add(time.Hour),
	}
	svc := NewService(store)

	_, err := svc.AcceptInvitation(context.Background(), raw, "u2", "other@example.com")
	if !errors.Is(err, ErrEmailMismatch) {
		t.Fatalf("err = %v, want ErrEmailMismatch", err)
	}
}

func TestRemoveMemberCannotRemoveOwner(t *testing.T) {
	t.Parallel()

	store := &stubStore{
		members: map[string]Membership{
			"w1|owner1": {ID: "m1", WorkspaceID: "w1", UserID: "owner1", Role: RoleOwner},
		},
	}
	svc := NewService(store)
	err := svc.RemoveMember(context.Background(), "w1", "owner1")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestChangeRoleRejectsOwnerTarget(t *testing.T) {
	t.Parallel()

	store := &stubStore{
		members: map[string]Membership{
			"w1|owner1": {ID: "m1", WorkspaceID: "w1", UserID: "owner1", Role: RoleOwner},
		},
	}
	svc := NewService(store)
	err := svc.ChangeRole(context.Background(), "w1", "owner1", RoleMember)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestLoadInvitationExpired(t *testing.T) {
	t.Parallel()

	store := &stubStore{invitations: map[string]Invitation{}}
	raw := "expired"
	store.invitations[hashInvitationToken(raw)] = Invitation{
		ID:        "inv1",
		Email:     "a@example.com",
		ExpiresAt: time.Now().UTC().Add(-time.Hour),
	}
	svc := NewService(store)
	_, err := svc.LoadInvitation(context.Background(), raw)
	if !errors.Is(err, ErrInvalidInvitation) {
		t.Fatalf("err = %v, want ErrInvalidInvitation", err)
	}
}
