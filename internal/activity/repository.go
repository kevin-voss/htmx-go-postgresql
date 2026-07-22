package activity

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store is the persistence port used by Service and RunAtomic.
type Store interface {
	Insert(ctx context.Context, e EventInput) (Event, error)
	InsertTx(ctx context.Context, tx Tx, e EventInput) (Event, error)
	ListByProject(ctx context.Context, projectID string, limit int) ([]Event, error)
	ListByWorkspace(ctx context.Context, workspaceID string, limit int) ([]Event, error)
}

type queryRower interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

// Repository persists activity events in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs a repository backed by pool.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const eventColumns = `e.id, e.workspace_id, COALESCE(e.project_id::text, ''), COALESCE(e.issue_id::text, ''),
	e.actor_id, COALESCE(u.display_name, ''), e.event_type, e.summary, e.created_at`

// Insert writes an activity event outside an explicit transaction.
func (r *Repository) Insert(ctx context.Context, e EventInput) (Event, error) {
	return r.insert(ctx, r.db, e)
}

// InsertTx writes an activity event using tx when it is a pgx transaction.
func (r *Repository) InsertTx(ctx context.Context, tx Tx, e EventInput) (Event, error) {
	pgxTx, ok := AsPgx(tx)
	if !ok {
		return Event{}, fmt.Errorf("activity: insert tx: unsupported transaction type")
	}
	return r.insert(ctx, pgxTx, e)
}

func (r *Repository) insert(ctx context.Context, q queryRower, e EventInput) (Event, error) {
	e.WorkspaceID = strings.TrimSpace(e.WorkspaceID)
	e.ProjectID = strings.TrimSpace(e.ProjectID)
	e.IssueID = strings.TrimSpace(e.IssueID)
	e.ActorID = strings.TrimSpace(e.ActorID)
	e.Type = strings.TrimSpace(e.Type)
	e.Summary = strings.TrimSpace(e.Summary)

	if e.WorkspaceID == "" || e.ActorID == "" || e.Type == "" || e.Summary == "" {
		return Event{}, fmt.Errorf("activity: insert: missing required fields")
	}

	const qInsert = `
		INSERT INTO activity_events (workspace_id, project_id, issue_id, actor_id, event_type, summary)
		VALUES (
			$1::uuid,
			NULLIF($2, '')::uuid,
			NULLIF($3, '')::uuid,
			$4::uuid,
			$5,
			$6
		)
		RETURNING id, workspace_id, COALESCE(project_id::text, ''), COALESCE(issue_id::text, ''),
			actor_id, event_type, summary, created_at`

	var event Event
	err := q.QueryRow(ctx, qInsert, e.WorkspaceID, e.ProjectID, e.IssueID, e.ActorID, e.Type, e.Summary).Scan(
		&event.ID,
		&event.WorkspaceID,
		&event.ProjectID,
		&event.IssueID,
		&event.ActorID,
		&event.Type,
		&event.Summary,
		&event.CreatedAt,
	)
	if err != nil {
		return Event{}, fmt.Errorf("activity: insert: %w", err)
	}
	return event, nil
}

// ListByProject returns recent events for a project, newest first.
func (r *Repository) ListByProject(ctx context.Context, projectID string, limit int) ([]Event, error) {
	if projectID == "" {
		return []Event{}, nil
	}
	if limit < 1 {
		limit = 20
	}
	const q = `
		SELECT ` + eventColumns + `
		FROM activity_events e
		LEFT JOIN users u ON u.id = e.actor_id
		WHERE e.project_id = $1::uuid
		ORDER BY e.created_at DESC, e.id DESC
		LIMIT $2`
	return r.list(ctx, q, projectID, limit)
}

// ListByWorkspace returns recent events for a workspace, newest first.
func (r *Repository) ListByWorkspace(ctx context.Context, workspaceID string, limit int) ([]Event, error) {
	if workspaceID == "" {
		return []Event{}, nil
	}
	if limit < 1 {
		limit = 20
	}
	const q = `
		SELECT ` + eventColumns + `
		FROM activity_events e
		LEFT JOIN users u ON u.id = e.actor_id
		WHERE e.workspace_id = $1::uuid
		ORDER BY e.created_at DESC, e.id DESC
		LIMIT $2`
	return r.list(ctx, q, workspaceID, limit)
}

func (r *Repository) list(ctx context.Context, q string, args ...any) ([]Event, error) {
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("activity: list: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(
			&e.ID,
			&e.WorkspaceID,
			&e.ProjectID,
			&e.IssueID,
			&e.ActorID,
			&e.ActorName,
			&e.Type,
			&e.Summary,
			&e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("activity: list scan: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("activity: list rows: %w", err)
	}
	if events == nil {
		events = []Event{}
	}
	return events, nil
}
