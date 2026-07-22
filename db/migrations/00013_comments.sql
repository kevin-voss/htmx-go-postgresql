-- +goose Up
CREATE TABLE issue_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id UUID NOT NULL REFERENCES issues (id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users (id),
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT issue_comments_body_len CHECK (char_length(body) BETWEEN 1 AND 10000)
);

CREATE INDEX issue_comments_issue_id_idx ON issue_comments (issue_id);
CREATE INDEX issue_comments_issue_id_created_at_idx ON issue_comments (issue_id, created_at);

-- +goose Down
DROP TABLE IF EXISTS issue_comments;
