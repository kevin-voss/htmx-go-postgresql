package workspace

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kevin-voss/htmx-go-postgresql/internal/member"
)

// ErrNotFound is returned when no workspace matches the lookup.
var ErrNotFound = errors.New("workspace: not found")

// ErrDuplicateSlug is returned when slug already exists.
var ErrDuplicateSlug = errors.New("workspace: duplicate slug")

// Repository persists workspaces in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs a repository backed by pool.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create inserts a workspace and Owner membership for createdBy in one transaction.
func (r *Repository) Create(ctx context.Context, name, slug, createdBy string) (Workspace, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Workspace{}, fmt.Errorf("workspace: create begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	const insertWorkspace = `
		INSERT INTO workspaces (name, slug, created_by)
		VALUES ($1, $2, $3)
		RETURNING id, name, slug, created_by, created_at, updated_at`

	var w Workspace
	err = tx.QueryRow(ctx, insertWorkspace, name, slug, createdBy).Scan(
		&w.ID,
		&w.Name,
		&w.Slug,
		&w.CreatedBy,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Workspace{}, ErrDuplicateSlug
		}
		return Workspace{}, fmt.Errorf("workspace: create: %w", err)
	}

	const insertOwner = `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, $3)`

	if _, err := tx.Exec(ctx, insertOwner, w.ID, createdBy, string(member.RoleOwner)); err != nil {
		return Workspace{}, fmt.Errorf("workspace: create owner membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Workspace{}, fmt.Errorf("workspace: create commit: %w", err)
	}
	return w, nil
}

// GetBySlug returns the workspace with the given slug.
func (r *Repository) GetBySlug(ctx context.Context, slug string) (Workspace, error) {
	const q = `
		SELECT id, name, slug, created_by, created_at, updated_at
		FROM workspaces
		WHERE slug = $1`

	var w Workspace
	err := r.db.QueryRow(ctx, q, slug).Scan(
		&w.ID,
		&w.Name,
		&w.Slug,
		&w.CreatedBy,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Workspace{}, ErrNotFound
		}
		return Workspace{}, fmt.Errorf("workspace: get by slug: %w", err)
	}
	return w, nil
}
