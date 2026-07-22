package issue

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrLabelNotFound is returned when no label matches the lookup.
var ErrLabelNotFound = errors.New("issue: label not found")

// ErrLabelExists is returned when a workspace already has a label with that name.
var ErrLabelExists = errors.New("issue: label already exists")

// ErrLabelNotInWorkspace is returned when attaching a label from another workspace.
var ErrLabelNotInWorkspace = errors.New("issue: label not in workspace")

const labelColumns = `id, workspace_id, name, color, created_at`

// CreateLabel inserts a workspace label.
func (r *Repository) CreateLabel(ctx context.Context, workspaceID, name, color string) (Label, error) {
	q := `
		INSERT INTO labels (workspace_id, name, color)
		VALUES ($1, $2, $3)
		RETURNING ` + labelColumns

	var label Label
	err := r.db.QueryRow(ctx, q, workspaceID, name, color).Scan(
		&label.ID,
		&label.WorkspaceID,
		&label.Name,
		&label.Color,
		&label.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return Label{}, ErrLabelExists
		}
		return Label{}, fmt.Errorf("issue: create label: %w", err)
	}
	return label, nil
}

// ListLabelsByWorkspace returns labels for a workspace ordered by name.
func (r *Repository) ListLabelsByWorkspace(ctx context.Context, workspaceID string) ([]Label, error) {
	q := `
		SELECT ` + labelColumns + `
		FROM labels
		WHERE workspace_id = $1
		ORDER BY lower(name) ASC`

	rows, err := r.db.Query(ctx, q, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("issue: list labels: %w", err)
	}
	defer rows.Close()

	var labels []Label
	for rows.Next() {
		label, err := scanLabel(rows)
		if err != nil {
			return nil, fmt.Errorf("issue: list labels scan: %w", err)
		}
		labels = append(labels, label)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("issue: list labels rows: %w", err)
	}
	if labels == nil {
		labels = []Label{}
	}
	return labels, nil
}

// GetLabelByID returns a label by id.
func (r *Repository) GetLabelByID(ctx context.Context, id string) (Label, error) {
	q := `SELECT ` + labelColumns + ` FROM labels WHERE id = $1`
	var label Label
	err := r.db.QueryRow(ctx, q, id).Scan(
		&label.ID,
		&label.WorkspaceID,
		&label.Name,
		&label.Color,
		&label.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Label{}, ErrLabelNotFound
		}
		return Label{}, fmt.Errorf("issue: get label: %w", err)
	}
	return label, nil
}

// DeleteLabel removes a workspace label (and its issue links via CASCADE).
func (r *Repository) DeleteLabel(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM labels WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("issue: delete label: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrLabelNotFound
	}
	return nil
}

// AttachLabel links a label to an issue. Idempotent if already attached.
func (r *Repository) AttachLabel(ctx context.Context, issueID, labelID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO issue_labels (issue_id, label_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING`, issueID, labelID)
	if err != nil {
		return fmt.Errorf("issue: attach label: %w", err)
	}
	return nil
}

// DetachLabel removes a label from an issue.
func (r *Repository) DetachLabel(ctx context.Context, issueID, labelID string) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM issue_labels
		WHERE issue_id = $1 AND label_id = $2`, issueID, labelID)
	if err != nil {
		return fmt.Errorf("issue: detach label: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrLabelNotFound
	}
	return nil
}

// ListLabelsForIssue returns labels attached to an issue.
func (r *Repository) ListLabelsForIssue(ctx context.Context, issueID string) ([]Label, error) {
	q := `
		SELECT l.id, l.workspace_id, l.name, l.color, l.created_at
		FROM labels l
		INNER JOIN issue_labels il ON il.label_id = l.id
		WHERE il.issue_id = $1
		ORDER BY lower(l.name) ASC`

	rows, err := r.db.Query(ctx, q, issueID)
	if err != nil {
		return nil, fmt.Errorf("issue: list labels for issue: %w", err)
	}
	defer rows.Close()

	var labels []Label
	for rows.Next() {
		label, err := scanLabel(rows)
		if err != nil {
			return nil, fmt.Errorf("issue: list labels for issue scan: %w", err)
		}
		labels = append(labels, label)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("issue: list labels for issue rows: %w", err)
	}
	if labels == nil {
		labels = []Label{}
	}
	return labels, nil
}

// ListLabelsForIssues returns attached labels keyed by issue id.
func (r *Repository) ListLabelsForIssues(ctx context.Context, issueIDs []string) (map[string][]Label, error) {
	out := make(map[string][]Label, len(issueIDs))
	if len(issueIDs) == 0 {
		return out, nil
	}

	q := `
		SELECT il.issue_id::text, l.id, l.workspace_id, l.name, l.color, l.created_at
		FROM issue_labels il
		INNER JOIN labels l ON l.id = il.label_id
		WHERE il.issue_id = ANY($1::uuid[])
		ORDER BY lower(l.name) ASC`

	rows, err := r.db.Query(ctx, q, issueIDs)
	if err != nil {
		return nil, fmt.Errorf("issue: list labels for issues: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var issueID string
		var label Label
		if err := rows.Scan(
			&issueID,
			&label.ID,
			&label.WorkspaceID,
			&label.Name,
			&label.Color,
			&label.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("issue: list labels for issues scan: %w", err)
		}
		out[issueID] = append(out[issueID], label)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("issue: list labels for issues rows: %w", err)
	}
	return out, nil
}

func scanLabel(row scannable) (Label, error) {
	var label Label
	err := row.Scan(
		&label.ID,
		&label.WorkspaceID,
		&label.Name,
		&label.Color,
		&label.CreatedAt,
	)
	return label, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	// Fallback for wrapped drivers / test doubles.
	return strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique")
}
