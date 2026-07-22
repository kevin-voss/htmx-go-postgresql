package member

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when membership (or workspace for access) is missing.
var ErrNotFound = errors.New("member: not found")

// ErrDuplicate is returned when the user is already a member of the workspace.
var ErrDuplicate = errors.New("member: duplicate membership")

// Repository persists workspace memberships and invitations in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs a repository backed by pool.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create inserts a membership row and returns the stored membership.
func (r *Repository) Create(ctx context.Context, workspaceID, userID string, role Role) (Membership, error) {
	const q = `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, workspace_id, user_id, role, created_at, updated_at`

	var m Membership
	var roleStr string
	err := r.db.QueryRow(ctx, q, workspaceID, userID, string(role)).Scan(
		&m.ID,
		&m.WorkspaceID,
		&m.UserID,
		&roleStr,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Membership{}, ErrDuplicate
		}
		return Membership{}, fmt.Errorf("member: create: %w", err)
	}
	m.Role = Role(roleStr)
	return m, nil
}

// GetByWorkspaceAndUser returns the membership for a user in a workspace.
func (r *Repository) GetByWorkspaceAndUser(ctx context.Context, workspaceID, userID string) (Membership, error) {
	const q = `
		SELECT id, workspace_id, user_id, role, created_at, updated_at
		FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2`

	var m Membership
	var roleStr string
	err := r.db.QueryRow(ctx, q, workspaceID, userID).Scan(
		&m.ID,
		&m.WorkspaceID,
		&m.UserID,
		&roleStr,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Membership{}, ErrNotFound
		}
		return Membership{}, fmt.Errorf("member: get by workspace and user: %w", err)
	}
	m.Role = Role(roleStr)
	return m, nil
}

// HasAny reports whether the user belongs to at least one workspace.
func (r *Repository) HasAny(ctx context.Context, userID string) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1 FROM workspace_members WHERE user_id = $1
		)`

	var ok bool
	if err := r.db.QueryRow(ctx, q, userID).Scan(&ok); err != nil {
		return false, fmt.Errorf("member: has any: %w", err)
	}
	return ok, nil
}

// GetAccessBySlug returns workspace + membership for a user by workspace slug.
// Missing workspace or membership both yield ErrNotFound (fail closed).
func (r *Repository) GetAccessBySlug(ctx context.Context, slug, userID string) (Access, error) {
	const q = `
		SELECT
			w.id, w.name, w.slug,
			m.id, m.workspace_id, m.user_id, m.role, m.created_at, m.updated_at
		FROM workspaces w
		INNER JOIN workspace_members m ON m.workspace_id = w.id
		WHERE w.slug = $1 AND m.user_id = $2`

	var a Access
	var roleStr string
	err := r.db.QueryRow(ctx, q, slug, userID).Scan(
		&a.WorkspaceID,
		&a.WorkspaceName,
		&a.WorkspaceSlug,
		&a.Membership.ID,
		&a.Membership.WorkspaceID,
		&a.Membership.UserID,
		&roleStr,
		&a.Membership.CreatedAt,
		&a.Membership.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Access{}, ErrNotFound
		}
		return Access{}, fmt.Errorf("member: get access by slug: %w", err)
	}
	a.Membership.Role = Role(roleStr)
	return a, nil
}

// ListByWorkspace returns memberships with user display fields for a workspace.
func (r *Repository) ListByWorkspace(ctx context.Context, workspaceID string) ([]MemberView, error) {
	const q = `
		SELECT
			m.id, m.workspace_id, m.user_id, m.role, m.created_at, m.updated_at,
			u.email, u.display_name
		FROM workspace_members m
		INNER JOIN users u ON u.id = m.user_id
		WHERE m.workspace_id = $1
		ORDER BY
			CASE m.role
				WHEN 'owner' THEN 0
				WHEN 'member' THEN 1
				ELSE 2
			END,
			u.display_name ASC`

	rows, err := r.db.Query(ctx, q, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("member: list by workspace: %w", err)
	}
	defer rows.Close()

	var out []MemberView
	for rows.Next() {
		var mv MemberView
		var roleStr string
		if err := rows.Scan(
			&mv.ID,
			&mv.WorkspaceID,
			&mv.UserID,
			&roleStr,
			&mv.CreatedAt,
			&mv.UpdatedAt,
			&mv.Email,
			&mv.DisplayName,
		); err != nil {
			return nil, fmt.Errorf("member: list by workspace scan: %w", err)
		}
		mv.Role = Role(roleStr)
		out = append(out, mv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("member: list by workspace rows: %w", err)
	}
	return out, nil
}

// UpdateRole sets a member's role. Owners cannot be demoted via this method
// (caller must enforce); DB allows any valid role string in workspace_members.
func (r *Repository) UpdateRole(ctx context.Context, workspaceID, userID string, role Role) error {
	const q = `
		UPDATE workspace_members
		SET role = $3, updated_at = now()
		WHERE workspace_id = $1 AND user_id = $2`

	tag, err := r.db.Exec(ctx, q, workspaceID, userID, string(role))
	if err != nil {
		return fmt.Errorf("member: update role: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a membership row.
func (r *Repository) Delete(ctx context.Context, workspaceID, userID string) error {
	const q = `
		DELETE FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2`

	tag, err := r.db.Exec(ctx, q, workspaceID, userID)
	if err != nil {
		return fmt.Errorf("member: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateInvitation inserts an invitation row (token hash only).
func (r *Repository) CreateInvitation(
	ctx context.Context,
	workspaceID, email string,
	role Role,
	invitedBy, tokenHash string,
	expiresAt time.Time,
) (Invitation, error) {
	const q = `
		INSERT INTO workspace_invitations (workspace_id, email, role, invited_by, token_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, workspace_id, email, role, invited_by, token_hash, created_at, expires_at, accepted_at`

	var inv Invitation
	var roleStr string
	err := r.db.QueryRow(ctx, q, workspaceID, email, string(role), invitedBy, tokenHash, expiresAt).Scan(
		&inv.ID,
		&inv.WorkspaceID,
		&inv.Email,
		&roleStr,
		&inv.InvitedBy,
		&inv.TokenHash,
		&inv.CreatedAt,
		&inv.ExpiresAt,
		&inv.AcceptedAt,
	)
	if err != nil {
		return Invitation{}, fmt.Errorf("member: create invitation: %w", err)
	}
	inv.Role = Role(roleStr)
	return inv, nil
}

// GetInvitationByTokenHash returns an invitation with workspace name/slug.
func (r *Repository) GetInvitationByTokenHash(ctx context.Context, tokenHash string) (Invitation, error) {
	const q = `
		SELECT
			i.id, i.workspace_id, i.email, i.role, i.invited_by, i.token_hash,
			i.created_at, i.expires_at, i.accepted_at,
			w.name, w.slug
		FROM workspace_invitations i
		INNER JOIN workspaces w ON w.id = i.workspace_id
		WHERE i.token_hash = $1`

	var inv Invitation
	var roleStr string
	err := r.db.QueryRow(ctx, q, tokenHash).Scan(
		&inv.ID,
		&inv.WorkspaceID,
		&inv.Email,
		&roleStr,
		&inv.InvitedBy,
		&inv.TokenHash,
		&inv.CreatedAt,
		&inv.ExpiresAt,
		&inv.AcceptedAt,
		&inv.WorkspaceName,
		&inv.WorkspaceSlug,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Invitation{}, ErrNotFound
		}
		return Invitation{}, fmt.Errorf("member: get invitation by token hash: %w", err)
	}
	inv.Role = Role(roleStr)
	return inv, nil
}

// AcceptInvitation creates membership and marks the invitation accepted in one transaction.
func (r *Repository) AcceptInvitation(ctx context.Context, invitationID, workspaceID, userID string, role Role, at time.Time) (Membership, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Membership{}, fmt.Errorf("member: accept invitation begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	const mark = `
		UPDATE workspace_invitations
		SET accepted_at = $2
		WHERE id = $1 AND accepted_at IS NULL AND expires_at > $2`
	tag, err := tx.Exec(ctx, mark, invitationID, at)
	if err != nil {
		return Membership{}, fmt.Errorf("member: mark invitation accepted: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Membership{}, ErrNotFound
	}

	const insert = `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, workspace_id, user_id, role, created_at, updated_at`

	var m Membership
	var roleStr string
	err = tx.QueryRow(ctx, insert, workspaceID, userID, string(role)).Scan(
		&m.ID,
		&m.WorkspaceID,
		&m.UserID,
		&roleStr,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Membership{}, ErrDuplicate
		}
		return Membership{}, fmt.Errorf("member: accept create membership: %w", err)
	}
	m.Role = Role(roleStr)

	if err := tx.Commit(ctx); err != nil {
		return Membership{}, fmt.Errorf("member: accept invitation commit: %w", err)
	}
	return m, nil
}

// MarkInvitationAccepted sets accepted_at when membership already exists (idempotent accept).
func (r *Repository) MarkInvitationAccepted(ctx context.Context, invitationID string, at time.Time) error {
	const q = `
		UPDATE workspace_invitations
		SET accepted_at = $2
		WHERE id = $1 AND accepted_at IS NULL`

	_, err := r.db.Exec(ctx, q, invitationID, at)
	if err != nil {
		return fmt.Errorf("member: mark invitation accepted: %w", err)
	}
	return nil
}
