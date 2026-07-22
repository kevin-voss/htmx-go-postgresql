-- +goose Up
CREATE TABLE labels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    color TEXT NOT NULL DEFAULT '#64748b',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT labels_name_len CHECK (char_length(name) BETWEEN 1 AND 40),
    CONSTRAINT labels_color_format CHECK (color ~ '^#[0-9A-Fa-f]{6}$')
);

CREATE UNIQUE INDEX labels_workspace_name_ci_key ON labels (workspace_id, lower(name));
CREATE INDEX labels_workspace_id_idx ON labels (workspace_id);

CREATE TABLE issue_labels (
    issue_id UUID NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES labels (id) ON DELETE CASCADE,
    PRIMARY KEY (issue_id, label_id)
);

CREATE INDEX issue_labels_label_id_idx ON issue_labels (label_id);

-- +goose Down
DROP TABLE IF EXISTS issue_labels;
DROP TABLE IF EXISTS labels;
