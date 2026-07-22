-- +goose Up
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT projects_workspace_slug_key UNIQUE (workspace_id, slug),
    CONSTRAINT projects_slug_lowercase CHECK (slug = lower(slug)),
    CONSTRAINT projects_slug_format CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    CONSTRAINT projects_slug_len CHECK (char_length(slug) BETWEEN 2 AND 48),
    CONSTRAINT projects_name_len CHECK (char_length(name) BETWEEN 2 AND 50)
);

CREATE INDEX projects_workspace_id_idx ON projects (workspace_id);

-- +goose Down
DROP TABLE IF EXISTS projects;
