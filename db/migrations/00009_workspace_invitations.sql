-- +goose Up
CREATE TABLE workspace_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',
    invited_by UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    CONSTRAINT workspace_invitations_token_hash_key UNIQUE (token_hash),
    CONSTRAINT workspace_invitations_role_check CHECK (role IN ('member', 'viewer'))
);

CREATE INDEX workspace_invitations_workspace_id_idx ON workspace_invitations (workspace_id);
CREATE INDEX workspace_invitations_email_idx ON workspace_invitations (email);
CREATE INDEX workspace_invitations_expires_at_idx ON workspace_invitations (expires_at);

-- +goose Down
DROP TABLE IF EXISTS workspace_invitations;
