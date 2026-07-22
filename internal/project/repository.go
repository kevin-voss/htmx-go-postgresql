package project

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when no project matches the lookup.
var ErrNotFound = errors.New("project: not found")

// ErrDuplicateSlug is returned when the project slug already exists in the workspace.
var ErrDuplicateSlug = errors.New("project: duplicate slug")

// Repository persists projects in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs a repository backed by pool.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create inserts a project row and returns the stored project.
func (r *Repository) Create(ctx context.Context, workspaceID, name, slug, createdBy string) (Project, error) {
	const q = `
		INSERT INTO projects (workspace_id, name, slug, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, workspace_id, name, slug, created_by, created_at, updated_at`

	var p Project
	err := r.db.QueryRow(ctx, q, workspaceID, name, slug, createdBy).Scan(
		&p.ID,
		&p.WorkspaceID,
		&p.Name,
		&p.Slug,
		&p.CreatedBy,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Project{}, ErrDuplicateSlug
		}
		return Project{}, fmt.Errorf("project: create: %w", err)
	}
	return p, nil
}

// ListByWorkspace returns projects for a workspace ordered by name.
func (r *Repository) ListByWorkspace(ctx context.Context, workspaceID string) ([]Project, error) {
	const q = `
		SELECT id, workspace_id, name, slug, created_by, created_at, updated_at
		FROM projects
		WHERE workspace_id = $1
		ORDER BY name ASC`

	rows, err := r.db.Query(ctx, q, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("project: list by workspace: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(
			&p.ID,
			&p.WorkspaceID,
			&p.Name,
			&p.Slug,
			&p.CreatedBy,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("project: list by workspace scan: %w", err)
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("project: list by workspace rows: %w", err)
	}
	if projects == nil {
		projects = []Project{}
	}
	return projects, nil
}

// GetByWorkspaceAndSlug returns a project by workspace id and project slug.
func (r *Repository) GetByWorkspaceAndSlug(ctx context.Context, workspaceID, slug string) (Project, error) {
	const q = `
		SELECT id, workspace_id, name, slug, created_by, created_at, updated_at
		FROM projects
		WHERE workspace_id = $1 AND slug = $2`

	var p Project
	err := r.db.QueryRow(ctx, q, workspaceID, slug).Scan(
		&p.ID,
		&p.WorkspaceID,
		&p.Name,
		&p.Slug,
		&p.CreatedBy,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Project{}, ErrNotFound
		}
		return Project{}, fmt.Errorf("project: get by workspace and slug: %w", err)
	}
	return p, nil
}
