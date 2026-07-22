package comment

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kevin-voss/htmx-go-postgresql/internal/activity"
)

// ErrNotFound is returned when no comment matches the lookup.
var ErrNotFound = errors.New("comment: not found")

// Repository persists issue comments in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs a repository backed by pool.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const commentColumns = `c.id, c.issue_id, c.author_id, u.display_name, c.body, c.created_at`

type queryRower interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

// Create inserts a comment and returns it with the author display name.
func (r *Repository) Create(ctx context.Context, issueID, authorID, body string) (Comment, error) {
	return r.createOn(ctx, r.db, issueID, authorID, body)
}

// CreateTx inserts a comment using an activity transaction started by the service layer.
func (r *Repository) CreateTx(ctx context.Context, tx activity.Tx, issueID, authorID, body string) (Comment, error) {
	pgxTx, ok := activity.AsPgx(tx)
	if !ok {
		return Comment{}, fmt.Errorf("comment: create tx: unsupported transaction type")
	}
	return r.createOn(ctx, pgxTx, issueID, authorID, body)
}

func (r *Repository) createOn(ctx context.Context, q queryRower, issueID, authorID, body string) (Comment, error) {
	const insert = `
		INSERT INTO issue_comments (issue_id, author_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, issue_id, author_id, body, created_at`

	var c Comment
	err := q.QueryRow(ctx, insert, issueID, authorID, body).Scan(
		&c.ID,
		&c.IssueID,
		&c.AuthorID,
		&c.Body,
		&c.CreatedAt,
	)
	if err != nil {
		return Comment{}, fmt.Errorf("comment: create: %w", err)
	}

	err = q.QueryRow(ctx, `SELECT display_name FROM users WHERE id = $1`, authorID).Scan(&c.AuthorName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.AuthorName = "Unknown"
			return c, nil
		}
		return Comment{}, fmt.Errorf("comment: author name: %w", err)
	}
	return c, nil
}

// ListByIssue returns comments for an issue ordered by creation time ascending.
func (r *Repository) ListByIssue(ctx context.Context, issueID string) ([]Comment, error) {
	const q = `
		SELECT ` + commentColumns + `
		FROM issue_comments c
		JOIN users u ON u.id = c.author_id
		WHERE c.issue_id = $1
		ORDER BY c.created_at ASC, c.id ASC`

	rows, err := r.db.Query(ctx, q, issueID)
	if err != nil {
		return nil, fmt.Errorf("comment: list by issue: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(
			&c.ID,
			&c.IssueID,
			&c.AuthorID,
			&c.AuthorName,
			&c.Body,
			&c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("comment: scan: %w", err)
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("comment: list rows: %w", err)
	}
	if comments == nil {
		comments = []Comment{}
	}
	return comments, nil
}

// CountByIssue returns the number of comments on an issue.
func (r *Repository) CountByIssue(ctx context.Context, issueID string) (int, error) {
	const q = `SELECT COUNT(*) FROM issue_comments WHERE issue_id = $1`
	var n int
	if err := r.db.QueryRow(ctx, q, issueID).Scan(&n); err != nil {
		return 0, fmt.Errorf("comment: count: %w", err)
	}
	return n, nil
}

// GetByID returns a comment by id, or ErrNotFound.
func (r *Repository) GetByID(ctx context.Context, id string) (Comment, error) {
	const q = `
		SELECT ` + commentColumns + `
		FROM issue_comments c
		JOIN users u ON u.id = c.author_id
		WHERE c.id = $1`

	var c Comment
	err := r.db.QueryRow(ctx, q, id).Scan(
		&c.ID,
		&c.IssueID,
		&c.AuthorID,
		&c.AuthorName,
		&c.Body,
		&c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Comment{}, ErrNotFound
		}
		return Comment{}, fmt.Errorf("comment: get by id: %w", err)
	}
	return c, nil
}

// Delete removes a comment by id.
func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM issue_comments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("comment: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
