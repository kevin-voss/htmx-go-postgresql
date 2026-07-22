# Step 31: Integration tests

| Field | Value |
| ----- | ----- |
| ID | `STEP-31` |
| Milestone | M6 — Portfolio quality |
| Status | `todo` |
| Depends on | STEP-30 |
| Unlocks | STEP-32 |
| Estimated scope | L |

---

## Goal

`make test` runs handler, authorization, and repository integration tests against Postgres.

## Description

Raise confidence for portfolio reviewers. Prefer tests/integration with test DB or Compose-run tests.

## References

- DoD: [DEFINITION_OF_DONE.md](../../DEFINITION_OF_DONE.md)
- Roles: [roles.md](../../specs/roles.md)
- Makefile: [makefile.md](../../architecture/makefile.md)

## Prerequisites

- Seed/fixtures helpful.
- Compose test path exists.

## Scope

### In

- Repository integration tests
- Authorization tests (cross-workspace denied)
- Key handler tests (httptest)
- make test green in CI-like environment

### Out

- 100% coverage obsession
- E2E browser suite

## Implementation checklist

- [ ] Add integration test helpers
- [ ] Cover authz + repo + at least one HTMX/handler path
- [ ] Ensure make test documented

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `tests/integration/**` | create |  |
| `internal/**/*_test.go` | create/modify |  |
| `Makefile` | modify | if needed |

## Technical notes

Tests must not require manual UI. Use testcontainers or compose run as already patterned.

## Acceptance criteria

- [ ] `make test` passes
- [ ] Authorization tests prove workspace isolation
- [ ] Repository integration tests hit real Postgres
- [ ] Handler tests cover at least login or issue create path

## Verification

```bash
make test
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
test(integration): add authz and repository integration tests
```

**Body:**

```text
Prove workspace isolation and persistence against Postgres so make test
is a credible quality gate.

STEP-31
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-32

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

How to run tests: make test. Special env vars: ____.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-32.
