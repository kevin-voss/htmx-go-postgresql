package member

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when membership (or workspace for access) is missing.
var ErrNotFound = errors.New("member: not found")

// ErrDuplicate is returned when the user is already a member of the workspace.
var ErrDuplicate = errors.New("member: duplicate membership")

// Repository persists workspace memberships in PostgreSQL.
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
