# Step 07: Database pool & goose migrations

| Field | Value |
| ----- | ----- |
| ID | `STEP-07` |
| Milestone | M1 — Foundation |
| Status | `todo` |
| Depends on | STEP-06 |
| Unlocks | STEP-08 |
| Estimated scope | M |

---

## Goal

App connects to PostgreSQL via pgx and can apply goose migrations; cmd/migrate exists.

## Description

Wire database connectivity and migration tooling. First migration may enable pgcrypto/uuid and create a trivial schema_migrations-friendly baseline; domain tables arrive in later steps.

## References

- Database: [database.md](../../architecture/database.md)
- Docker: [docker.md](../../architecture/docker.md)

## Prerequisites

- Compose Postgres healthy.
- DATABASE_URL configured.

## Scope

### In

- pgx pool in internal/database
- goose migrations under db/migrations
- cmd/migrate up/down
- App fails fast if DB unreachable at startup (or retries briefly)
- Optional: /health can report DB status — nice-to-have

### Out

- users/sessions tables (auth steps)
- sqlc generation (optional later)

## Implementation checklist

- [ ] Add pgx dependency
- [ ] Implement pool open/close
- [ ] Add goose + first migration (uuid extension)
- [ ] cmd/migrate works against Compose DB

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `internal/database/database.go` | create | pool |
| `db/migrations/*.sql` | create | initial migration |
| `cmd/migrate/main.go` | create | goose runner |
| `go.mod` | modify | deps |

## Technical notes

Use gen_random_uuid() — ensure pgcrypto or built-in uuid on Postgres 18. Prefer goose SQL migrations.

## Acceptance criteria

- [ ] App opens a pgx pool successfully against Compose DB
- [ ] `go run ./cmd/migrate up` applies migrations cleanly
- [ ] Re-running migrate up is idempotent (no error)
- [ ] Pool closes on shutdown

## Verification

```bash
docker compose up -d database
go run ./cmd/migrate up
go test ./internal/database/...
```

## Commit

**Subject (required):**

```text
feat(step-07): connect PostgreSQL and add goose migrations
```

**Body (optional):**

```text
Complete STEP-07 so the next agent can continue from a green tree.
```

## Handoff to next agent

Migration naming scheme: ____. Pool injected into Application.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-08.
