-- +goose Up
CREATE TABLE issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    issue_number INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'backlog',
    created_by UUID NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT issues_project_number_key UNIQUE (project_id, issue_number),
    CONSTRAINT issues_issue_number_positive CHECK (issue_number > 0),
    CONSTRAINT issues_title_len CHECK (char_length(title) BETWEEN 1 AND 200),
    CONSTRAINT issues_status_check CHECK (status IN ('backlog', 'todo', 'in_progress', 'done'))
);

CREATE INDEX issues_project_id_idx ON issues (project_id);
CREATE INDEX issues_project_id_number_idx ON issues (project_id, issue_number);

-- +goose Down
DROP TABLE IF EXISTS issues;
