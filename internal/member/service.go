package member

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
)

// ErrInvalidInvitation is returned for missing, used, or expired invitations.
var ErrInvalidInvitation = errors.New("member: invalid invitation")

// ErrEmailMismatch is returned when the signed-in user email does not match the invite.
var ErrEmailMismatch = errors.New("member: invitation email mismatch")

// ErrForbidden is returned when a role-management action is not allowed.
var ErrForbidden = errors.New("member: forbidden")

// ErrAlreadyMember is returned when inviting an email that already has membership.
var ErrAlreadyMember = errors.New("member: already a member")

// Store is the persistence port used by Service.
type Store interface {
	Create(ctx context.Context, workspaceID, userID string, role Role) (Membership, error)
	GetByWorkspaceAndUser(ctx context.Context, workspaceID, userID string) (Membership, error)
	GetAccessBySlug(ctx context.Context, slug, userID string) (Access, error)
	HasAny(ctx context.Context, userID string) (bool, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]MemberView, error)
	UpdateRole(ctx context.Context, workspaceID, userID string, role Role) error
	Delete(ctx context.Context, workspaceID, userID string) error
	CreateInvitation(ctx context.Context, workspaceID, email string, role Role, invitedBy, tokenHash string, expiresAt time.Time) (Invitation, error)
	GetInvitationByTokenHash(ctx context.Context, tokenHash string) (Invitation, error)
	AcceptInvitation(ctx context.Context, invitationID, workspaceID, userID string, role Role, at time.Time) (Membership, error)
	MarkInvitationAccepted(ctx context.Context, invitationID string, at time.Time) error
}

// UserEmailLookup finds a user id by email for invite checks.
type UserEmailLookup interface {
	GetUserIDByEmail(ctx context.Context, email string) (string, error)
}

// Service implements membership lookups, invitations, and role helpers.
type Service struct {
	store Store
	users UserEmailLookup
}

// NewService constructs a membership service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// WithUserLookup attaches optional email→user lookup (for invite duplicate checks).
func (s *Service) WithUserLookup(users UserEmailLookup) *Service {
	s.users = users
	return s
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

// ListMembers returns workspace members for the members page.
func (s *Service) ListMembers(ctx context.Context, workspaceID string) ([]MemberView, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return nil, ErrNotFound
	}
	return s.store.ListByWorkspace(ctx, workspaceID)
}

// InviteErrors holds per-field validation messages for the invite form.
type InviteErrors struct {
	Email string
	Role  string
}

// Any reports whether any invite field error is set.
func (e InviteErrors) Any() bool {
	return e.Email != "" || e.Role != ""
}

// CreateInvitationInput is the owner invite form payload.
type CreateInvitationInput struct {
	WorkspaceID   string
	WorkspaceName string
	Email         string
	Role          Role
	InvitedBy     string
}

// CreateInvitation validates input, persists a hashed token, and returns the raw token.
func (s *Service) CreateInvitation(ctx context.Context, in CreateInvitationInput) (rawToken string, fieldErrs InviteErrors, err error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	role := in.Role
	if role == "" {
		role = RoleMember
	}

	if email == "" || !validEmail(email) {
		fieldErrs.Email = "Enter a valid email address."
	}
	if !role.Invitable() {
		fieldErrs.Role = "Choose Member or Viewer."
	}
	if fieldErrs.Any() {
		return "", fieldErrs, nil
	}

	if s.users != nil {
		userID, lookupErr := s.users.GetUserIDByEmail(ctx, email)
		if lookupErr == nil && userID != "" {
			_, memErr := s.store.GetByWorkspaceAndUser(ctx, in.WorkspaceID, userID)
			if memErr == nil {
				fieldErrs.Email = "That user is already a member of this workspace."
				return "", fieldErrs, nil
			}
			if !errors.Is(memErr, ErrNotFound) {
				return "", InviteErrors{}, memErr
			}
		} else if lookupErr != nil && !errors.Is(lookupErr, ErrNotFound) {
			return "", InviteErrors{}, lookupErr
		}
	}

	rawToken, err = generateInvitationToken()
	if err != nil {
		return "", InviteErrors{}, err
	}
	expiresAt := time.Now().UTC().Add(invitationTTL)
	_, err = s.store.CreateInvitation(
		ctx,
		in.WorkspaceID,
		email,
		role,
		in.InvitedBy,
		hashInvitationToken(rawToken),
		expiresAt,
	)
	if err != nil {
		return "", InviteErrors{}, err
	}
	return rawToken, InviteErrors{}, nil
}

// LoadInvitation returns a valid pending invitation for the raw token.
func (s *Service) LoadInvitation(ctx context.Context, rawToken string) (Invitation, error) {
	if strings.TrimSpace(rawToken) == "" {
		return Invitation{}, ErrInvalidInvitation
	}
	inv, err := s.store.GetInvitationByTokenHash(ctx, hashInvitationToken(rawToken))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Invitation{}, ErrInvalidInvitation
		}
		return Invitation{}, err
	}
	now := time.Now().UTC()
	if inv.AcceptedAt != nil || !inv.ExpiresAt.After(now) {
		return Invitation{}, ErrInvalidInvitation
	}
	return inv, nil
}

// AcceptInvitationResult describes the outcome of accepting an invitation.
type AcceptInvitationResult struct {
	Invitation Invitation
	Membership Membership
	Already    bool
}

// AcceptInvitation creates membership for the signed-in user matching the invite email.
func (s *Service) AcceptInvitation(ctx context.Context, rawToken, userID, userEmail string) (AcceptInvitationResult, error) {
	inv, err := s.LoadInvitation(ctx, rawToken)
	if err != nil {
		return AcceptInvitationResult{}, err
	}

	normalized := strings.ToLower(strings.TrimSpace(userEmail))
	if normalized == "" || normalized != inv.Email {
		return AcceptInvitationResult{}, ErrEmailMismatch
	}

	existing, err := s.store.GetByWorkspaceAndUser(ctx, inv.WorkspaceID, userID)
	if err == nil {
		_ = s.store.MarkInvitationAccepted(ctx, inv.ID, time.Now().UTC())
		return AcceptInvitationResult{Invitation: inv, Membership: existing, Already: true}, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return AcceptInvitationResult{}, err
	}

	now := time.Now().UTC()
	m, err := s.store.AcceptInvitation(ctx, inv.ID, inv.WorkspaceID, userID, inv.Role, now)
	if err != nil {
		if errors.Is(err, ErrDuplicate) {
			existing, getErr := s.store.GetByWorkspaceAndUser(ctx, inv.WorkspaceID, userID)
			if getErr != nil {
				return AcceptInvitationResult{}, getErr
			}
			_ = s.store.MarkInvitationAccepted(ctx, inv.ID, now)
			return AcceptInvitationResult{Invitation: inv, Membership: existing, Already: true}, nil
		}
		if errors.Is(err, ErrNotFound) {
			return AcceptInvitationResult{}, ErrInvalidInvitation
		}
		return AcceptInvitationResult{}, err
	}
	return AcceptInvitationResult{Invitation: inv, Membership: m, Already: false}, nil
}

// ChangeRole updates a non-owner member to Member or Viewer.
func (s *Service) ChangeRole(ctx context.Context, workspaceID, targetUserID string, newRole Role) error {
	if !newRole.Invitable() {
		return fmt.Errorf("member: invalid role %q: %w", newRole, ErrForbidden)
	}
	target, err := s.store.GetByWorkspaceAndUser(ctx, workspaceID, targetUserID)
	if err != nil {
		return err
	}
	if target.Role == RoleOwner {
		return ErrForbidden
	}
	if target.Role == newRole {
		return nil
	}
	return s.store.UpdateRole(ctx, workspaceID, targetUserID, newRole)
}

// RemoveMember removes a non-owner member from the workspace.
func (s *Service) RemoveMember(ctx context.Context, workspaceID, targetUserID string) error {
	target, err := s.store.GetByWorkspaceAndUser(ctx, workspaceID, targetUserID)
	if err != nil {
		return err
	}
	if target.Role == RoleOwner {
		return ErrForbidden
	}
	return s.store.Delete(ctx, workspaceID, targetUserID)
}

func validEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return addr.Address == email && strings.Contains(email, "@")
}
