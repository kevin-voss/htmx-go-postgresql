package issue

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when no issue matches the lookup.
var ErrNotFound = errors.New("issue: not found")

// Repository persists issues in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs a repository backed by pool.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create inserts an issue with the next per-project issue_number.
// Allocation locks the project row so concurrent creates cannot collide.
func (r *Repository) Create(ctx context.Context, projectID, title, description, createdBy string) (Issue, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Issue{}, fmt.Errorf("issue: begin create: %w", err)
	}
	defer tx.Rollback(ctx)

	var lockedProjectID string
	err = tx.QueryRow(ctx, `SELECT id FROM projects WHERE id = $1 FOR UPDATE`, projectID).Scan(&lockedProjectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Issue{}, ErrNotFound
		}
		return Issue{}, fmt.Errorf("issue: lock project: %w", err)
	}

	var next int
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(issue_number), 0) + 1
		FROM issues
		WHERE project_id = $1`, projectID).Scan(&next)
	if err != nil {
		return Issue{}, fmt.Errorf("issue: next number: %w", err)
	}

	const insert = `
		INSERT INTO issues (project_id, issue_number, title, description, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, project_id, issue_number, title, description, status, created_by, created_at, updated_at`

	var issue Issue
	err = tx.QueryRow(ctx, insert, projectID, next, title, description, StatusBacklog, createdBy).Scan(
		&issue.ID,
		&issue.ProjectID,
		&issue.IssueNumber,
		&issue.Title,
		&issue.Description,
		&issue.Status,
		&issue.CreatedBy,
		&issue.CreatedAt,
		&issue.UpdatedAt,
	)
	if err != nil {
		return Issue{}, fmt.Errorf("issue: insert: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Issue{}, fmt.Errorf("issue: commit create: %w", err)
	}
	return issue, nil
}

// ListByProject returns issues for a project ordered by issue_number ascending.
func (r *Repository) ListByProject(ctx context.Context, projectID string) ([]Issue, error) {
	const q = `
		SELECT id, project_id, issue_number, title, description, status, created_by, created_at, updated_at
		FROM issues
		WHERE project_id = $1
		ORDER BY issue_number ASC`

	rows, err := r.db.Query(ctx, q, projectID)
	if err != nil {
		return nil, fmt.Errorf("issue: list by project: %w", err)
	}
	defer rows.Close()

	var issues []Issue
	for rows.Next() {
		var issue Issue
		if err := rows.Scan(
			&issue.ID,
			&issue.ProjectID,
			&issue.IssueNumber,
			&issue.Title,
			&issue.Description,
			&issue.Status,
			&issue.CreatedBy,
			&issue.CreatedAt,
			&issue.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("issue: list by project scan: %w", err)
		}
		issues = append(issues, issue)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("issue: list by project rows: %w", err)
	}
	if issues == nil {
		issues = []Issue{}
	}
	return issues, nil
}

// GetByProjectAndNumber returns an issue by project id and issue number.
func (r *Repository) GetByProjectAndNumber(ctx context.Context, projectID string, issueNumber int) (Issue, error) {
	const q = `
		SELECT id, project_id, issue_number, title, description, status, created_by, created_at, updated_at
		FROM issues
		WHERE project_id = $1 AND issue_number = $2`

	var issue Issue
	err := r.db.QueryRow(ctx, q, projectID, issueNumber).Scan(
		&issue.ID,
		&issue.ProjectID,
		&issue.IssueNumber,
		&issue.Title,
		&issue.Description,
		&issue.Status,
		&issue.CreatedBy,
		&issue.CreatedAt,
		&issue.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Issue{}, ErrNotFound
		}
		return Issue{}, fmt.Errorf("issue: get by project and number: %w", err)
	}
	return issue, nil
}

// GetByWorkspaceAndNumber returns an issue by workspace id and issue number.
// When multiple projects share the same number, returns ErrNotFound (ambiguous).
func (r *Repository) GetByWorkspaceAndNumber(ctx context.Context, workspaceID string, issueNumber int) (Issue, error) {
	const q = `
		SELECT i.id, i.project_id, i.issue_number, i.title, i.description, i.status,
			i.created_by, i.created_at, i.updated_at
		FROM issues i
		INNER JOIN projects p ON p.id = i.project_id
		WHERE p.workspace_id = $1 AND i.issue_number = $2
		ORDER BY p.slug ASC`

	rows, err := r.db.Query(ctx, q, workspaceID, issueNumber)
	if err != nil {
		return Issue{}, fmt.Errorf("issue: get by workspace and number: %w", err)
	}
	defer rows.Close()

	var matches []Issue
	for rows.Next() {
		var issue Issue
		if err := rows.Scan(
			&issue.ID,
			&issue.ProjectID,
			&issue.IssueNumber,
			&issue.Title,
			&issue.Description,
			&issue.Status,
			&issue.CreatedBy,
			&issue.CreatedAt,
			&issue.UpdatedAt,
		); err != nil {
			return Issue{}, fmt.Errorf("issue: get by workspace and number scan: %w", err)
		}
		matches = append(matches, issue)
	}
	if err := rows.Err(); err != nil {
		return Issue{}, fmt.Errorf("issue: get by workspace and number rows: %w", err)
	}
	if len(matches) != 1 {
		return Issue{}, ErrNotFound
	}
	return matches[0], nil
}
