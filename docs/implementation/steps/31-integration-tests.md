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

## Commit

**Subject (required):**

```text
test(step-31): add integration tests for authz and repositories
```

**Body (optional):**

```text
Complete STEP-31 so the next agent can continue from a green tree.
```

## Handoff to next agent

How to run tests: make test. Special env vars: ____.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-32.
