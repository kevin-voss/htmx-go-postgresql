-- +goose Up
ALTER TABLE issues
    ADD COLUMN priority TEXT NOT NULL DEFAULT 'medium',
    ADD COLUMN assignee_id UUID REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN archived BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE issues
    ADD CONSTRAINT issues_priority_check
    CHECK (priority IN ('low', 'medium', 'high', 'urgent'));

CREATE INDEX issues_project_id_active_idx ON issues (project_id) WHERE archived = false;
CREATE INDEX issues_assignee_id_idx ON issues (assignee_id) WHERE assignee_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS issues_assignee_id_idx;
DROP INDEX IF EXISTS issues_project_id_active_idx;

ALTER TABLE issues DROP CONSTRAINT IF EXISTS issues_priority_check;
ALTER TABLE issues
    DROP COLUMN IF EXISTS archived,
    DROP COLUMN IF EXISTS assignee_id,
    DROP COLUMN IF EXISTS priority;
