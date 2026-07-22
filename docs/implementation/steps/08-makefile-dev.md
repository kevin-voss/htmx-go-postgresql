# Step 08: Makefile & make dev

| Field | Value |
| ----- | ----- |
| ID | `STEP-08` |
| Milestone | M1 — Foundation |
| Status | `todo` |
| Depends on | STEP-07 |
| Unlocks | STEP-09 |
| Estimated scope | S |

---

## Goal

`make dev` builds/starts the stack; development entrypoint runs migrations then the web app.

## Description

Make the clone→make dev promise real. Align Makefile targets with architecture/makefile.md.

## References

- Makefile: [makefile.md](../../architecture/makefile.md)
- Docker: [docker.md](../../architecture/docker.md)
- Definition of done: [DEFINITION_OF_DONE.md](../../DEFINITION_OF_DONE.md)

## Prerequisites

- Dockerfile + compose + migrate + web server exist.

## Scope

### In

- Makefile: help, dev, stop, reset, test, lint, migrate, seed(stub), logs
- Dev entrypoint script: migrate up && web
- Document URLs in comment or brief README stub

### Out

- Seed implementation (step 30)
- Full portfolio README (step 32)

## Implementation checklist

- [ ] Add Makefile
- [ ] Wire entrypoint
- [ ] make dev brings app to :8080
- [ ] make stop / make migrate work

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `Makefile` | create | targets |
| `scripts/dev-entrypoint.sh or Dockerfile CMD` | create/modify | migrate then web |
| `README.md` | create/modify | minimal clone instructions OK |

## Technical notes

Keep .DEFAULT_GOAL := help. Do not require npm.

## Acceptance criteria

- [ ] `make dev` eventually serves GET /health on :8080
- [ ] Migrations run automatically on app container start
- [ ] `make stop` tears down containers
- [ ] help target lists commands

## Verification

```bash
make help
make dev
# then curl localhost:8080/health
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
chore(dx): add Makefile and migrate-then-run entrypoint
```

**Body:**

```text
Make clone-to-running a single make dev flow, including automatic
migrations before the web process starts.

STEP-08
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-09

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Primary DX command is make dev. Seed target may stub until step 30.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-09.
