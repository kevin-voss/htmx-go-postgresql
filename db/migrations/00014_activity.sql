-- +goose Up
CREATE TABLE activity_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects (id) ON DELETE CASCADE,
    issue_id UUID REFERENCES issues (id) ON DELETE SET NULL,
    actor_id UUID NOT NULL REFERENCES users (id),
    event_type TEXT NOT NULL,
    summary TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT activity_events_type_check CHECK (
        event_type IN ('issue.created', 'issue.status_changed', 'comment.created')
    ),
    CONSTRAINT activity_events_summary_len CHECK (char_length(summary) BETWEEN 1 AND 500)
);

CREATE INDEX activity_events_workspace_id_created_at_idx
    ON activity_events (workspace_id, created_at DESC);

CREATE INDEX activity_events_project_id_created_at_idx
    ON activity_events (project_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS activity_events;
