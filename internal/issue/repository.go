package issue

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

const issueColumns = `id, project_id, issue_number, title, description, status, priority,
	COALESCE(assignee_id::text, ''), archived, created_by, created_at, updated_at`

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

	insert := `
		INSERT INTO issues (project_id, issue_number, title, description, status, priority, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING ` + issueColumns

	var issue Issue
	err = tx.QueryRow(ctx, insert, projectID, next, title, description, StatusBacklog, PriorityMedium, createdBy).Scan(
		&issue.ID,
		&issue.ProjectID,
		&issue.IssueNumber,
		&issue.Title,
		&issue.Description,
		&issue.Status,
		&issue.Priority,
		&issue.AssigneeID,
		&issue.Archived,
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

// ListByProject returns non-archived issues for a project ordered by issue_number ascending.
// Optional filter fields combine with AND; empty fields are ignored.
func (r *Repository) ListByProject(ctx context.Context, projectID string, filter ListFilter) ([]Issue, error) {
	filter = NormalizeListFilter(filter)

	args := []any{projectID}
	conditions := []string{"project_id = $1", "archived = false"}
	n := 2

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", n))
		args = append(args, filter.Status)
		n++
	}
	if filter.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("priority = $%d", n))
		args = append(args, filter.Priority)
		n++
	}
	if filter.AssigneeID != "" {
		if filter.AssigneeID == "none" {
			conditions = append(conditions, "assignee_id IS NULL")
		} else {
			conditions = append(conditions, fmt.Sprintf("assignee_id = $%d::uuid", n))
			args = append(args, filter.AssigneeID)
			n++
		}
	}
	if filter.LabelID != "" {
		conditions = append(conditions, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM issue_labels il
			WHERE il.issue_id = issues.id AND il.label_id = $%d::uuid
		)`, n))
		args = append(args, filter.LabelID)
		n++
	}
	if filter.Query != "" {
		pattern := "%" + escapeLike(filter.Query) + "%"
		conditions = append(conditions, fmt.Sprintf(
			`(title ILIKE $%d ESCAPE '\' OR description ILIKE $%d ESCAPE '\')`, n, n,
		))
		args = append(args, pattern)
	}

	q := `
		SELECT ` + issueColumns + `
		FROM issues
		WHERE ` + strings.Join(conditions, " AND ") + `
		ORDER BY issue_number ASC`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("issue: list by project: %w", err)
	}
	defer rows.Close()

	var issues []Issue
	for rows.Next() {
		issue, err := scanIssue(rows)
		if err != nil {
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
	q := `
		SELECT ` + issueColumns + `
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
		&issue.Priority,
		&issue.AssigneeID,
		&issue.Archived,
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
	q := `
		SELECT i.id, i.project_id, i.issue_number, i.title, i.description, i.status, i.priority,
			COALESCE(i.assignee_id::text, ''), i.archived, i.created_by, i.created_at, i.updated_at
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
		issue, err := scanIssue(rows)
		if err != nil {
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

// UpdateStatus sets the status of an issue by id.
func (r *Repository) UpdateStatus(ctx context.Context, id, status string) (Issue, error) {
	return r.updateField(ctx, id, `status = $2`, status)
}

// UpdatePriority sets the priority of an issue by id.
func (r *Repository) UpdatePriority(ctx context.Context, id, priority string) (Issue, error) {
	return r.updateField(ctx, id, `priority = $2`, priority)
}

// UpdateAssignee sets or clears the assignee of an issue by id.
// Empty assigneeID clears the assignment.
func (r *Repository) UpdateAssignee(ctx context.Context, id, assigneeID string) (Issue, error) {
	if assigneeID == "" {
		q := `
			UPDATE issues
			SET assignee_id = NULL, updated_at = now()
			WHERE id = $1
			RETURNING ` + issueColumns
		return r.scanOne(ctx, q, id)
	}
	return r.updateField(ctx, id, `assignee_id = $2::uuid`, assigneeID)
}

// Archive marks an issue as archived (soft flag).
func (r *Repository) Archive(ctx context.Context, id string) (Issue, error) {
	q := `
		UPDATE issues
		SET archived = true, updated_at = now()
		WHERE id = $1
		RETURNING ` + issueColumns
	return r.scanOne(ctx, q, id)
}

func (r *Repository) updateField(ctx context.Context, id, setExpr, value string) (Issue, error) {
	q := `
		UPDATE issues
		SET ` + setExpr + `, updated_at = now()
		WHERE id = $1
		RETURNING ` + issueColumns
	return r.scanOne(ctx, q, id, value)
}

func (r *Repository) scanOne(ctx context.Context, q string, args ...any) (Issue, error) {
	var issue Issue
	err := r.db.QueryRow(ctx, q, args...).Scan(
		&issue.ID,
		&issue.ProjectID,
		&issue.IssueNumber,
		&issue.Title,
		&issue.Description,
		&issue.Status,
		&issue.Priority,
		&issue.AssigneeID,
		&issue.Archived,
		&issue.CreatedBy,
		&issue.CreatedAt,
		&issue.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Issue{}, ErrNotFound
		}
		return Issue{}, fmt.Errorf("issue: update: %w", err)
	}
	return issue, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanIssue(row scannable) (Issue, error) {
	var issue Issue
	err := row.Scan(
		&issue.ID,
		&issue.ProjectID,
		&issue.IssueNumber,
		&issue.Title,
		&issue.Description,
		&issue.Status,
		&issue.Priority,
		&issue.AssigneeID,
		&issue.Archived,
		&issue.CreatedBy,
		&issue.CreatedAt,
		&issue.UpdatedAt,
	)
	return issue, err
}
