# Step 02: Docker Compose services

| Field | Value |
| ----- | ----- |
| ID | `STEP-02` |
| Milestone | M1 — Foundation |
| Status | `todo` |
| Depends on | STEP-01 |
| Unlocks | STEP-03 |
| Estimated scope | M |

---

## Goal

Local Docker Compose brings up Postgres and Mailpit (and a development app image shell) matching the architecture doc.

## Description

Introduce containerized dependencies. The app service may still be a thin development image that can compile/run Go; full entrypoint wiring lands in step 08. Follow compose.yaml from architecture/docker.md closely.

## References

- Docker: [docker.md](../../architecture/docker.md)
- Makefile (later): [makefile.md](../../architecture/makefile.md)

## Prerequisites

- STEP-01 done.
- Docker Desktop / Engine available.

## Scope

### In

- Dockerfile with development target
- compose.yaml: app, database (postgres:18-alpine), mailpit
- Healthcheck on Postgres
- Named volume for DB data
- Document ports 8080 / 5432 / 8025 / 1025

### Out

- Automatic migrations on start (step 08).
- Production multi-stage finalize (step 32).

## Implementation checklist

- [ ] Write Dockerfile development stage
- [ ] Write compose.yaml per architecture doc
- [ ] docker compose config validates
- [ ] database and mailpit start healthy

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `Dockerfile` | create | development target |
| `compose.yaml` | create | app + database + mailpit |

## Technical notes

Use forgeboard/forgeboard credentials for local only — never for production. App may fail to listen until step 04; that is OK if DB/Mailpit are healthy.

## Acceptance criteria

- [ ] `docker compose config` succeeds
- [ ] Postgres healthcheck becomes healthy
- [ ] Mailpit UI reachable at http://localhost:8025
- [ ] Compose env vars match .env.example names

## Verification

```bash
docker compose config
docker compose up -d database mailpit
curl -sf http://localhost:8025 >/dev/null
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
chore(docker): add Compose for app, Postgres, and Mailpit
```

**Body:**

```text
Give every later step a shared local stack for the database and
development email without installing services on the host.

STEP-02
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-03

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Compose file path is compose.yaml. App service may not serve HTTP yet.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-03.
