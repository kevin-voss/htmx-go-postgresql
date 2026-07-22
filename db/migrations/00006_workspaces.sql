-- +goose Up
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT workspaces_slug_key UNIQUE (slug),
    CONSTRAINT workspaces_slug_lowercase CHECK (slug = lower(slug)),
    CONSTRAINT workspaces_slug_format CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    CONSTRAINT workspaces_slug_len CHECK (char_length(slug) BETWEEN 2 AND 48),
    CONSTRAINT workspaces_name_len CHECK (char_length(name) BETWEEN 2 AND 50)
);

-- +goose Down
DROP TABLE IF EXISTS workspaces;
