-- +goose Up
-- Enable pgcrypto so gen_random_uuid() is available for UUID primary keys.
-- PostgreSQL 13+ also ships gen_random_uuid() in core; the extension is
-- kept as an explicit baseline for domain migrations.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- +goose Down
DROP EXTENSION IF EXISTS pgcrypto;
