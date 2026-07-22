# Step 30: Seed command & demo account

| Field | Value |
| ----- | ----- |
| ID | `STEP-30` |
| Milestone | M6 — Portfolio quality |
| Status | `todo` |
| Depends on | STEP-29 |
| Unlocks | STEP-31 |
| Estimated scope | M |

---

## Goal

`make seed` inserts a demo account, workspace, project, issues, comments, and activity suitable for portfolio demos.

## Description

Deterministic seed data so reviewers need not click endlessly. Document credentials in README later (step 32) — put them in command output and .env.example comments now.

## References

- Makefile: [makefile.md](../../architecture/makefile.md)
- DoD: [DEFINITION_OF_DONE.md](../../DEFINITION_OF_DONE.md)

## Prerequisites

- Domain tables exist.

## Scope

### In

- cmd/seed
- Idempotent or reset-safe seeding strategy (document)
- Demo user + workspace + project + ≥N issues + comments + labels
- make seed target works

### Out

- Production data migrations of demo content

## Implementation checklist

- [ ] Implement seed
- [ ] Wire make seed
- [ ] Document demo credentials in seed output

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `cmd/seed/main.go` | create |  |
| `tests/fixtures/**` | create | optional shared fixtures |
| `Makefile` | modify | seed target |

## Technical notes

Use a well-known demo password only for local/dev. Never use it as a default in production builds.

## Acceptance criteria

- [ ] `make seed` succeeds against local DB
- [ ] Demo user can log in
- [ ] Seeded workspace shows multiple issues and activity
- [ ] Running seed does not corrupt unrelated prod assumptions (dev-oriented)

## Verification

```bash
make seed
# login as demo user
```

## Commit

**Subject (required):**

```text
feat(step-30): add seed command and demo account
```

**Body (optional):**

```text
Complete STEP-30 so the next agent can continue from a green tree.
```

## Handoff to next agent

Demo email/password: ____. Idempotency: ____.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-31.
