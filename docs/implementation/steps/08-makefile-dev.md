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

## Commit

**Subject (required):**

```text
chore(step-08): add Makefile and development entrypoint
```

**Body (optional):**

```text
Complete STEP-08 so the next agent can continue from a green tree.
```

## Handoff to next agent

Primary DX command is make dev. Seed target may stub until step 30.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-09.
