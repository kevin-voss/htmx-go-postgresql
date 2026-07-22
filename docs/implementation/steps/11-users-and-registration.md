# Step 11: Users & registration

| Field | Value |
| ----- | ----- |
| ID | `STEP-11` |
| Milestone | M2 — Authentication |
| Status | `todo` |
| Depends on | STEP-10 |
| Unlocks | STEP-12 |
| Estimated scope | L |

---

## Goal

Users can register via GET/POST /register with validation; user rows persist; session creation may wait for step 12 if explicitly deferred — prefer creating session if sessions table already planned, else redirect stub.

## Description

Implement registration flow fields and validation from the registration flow doc. Email verification sending can stub until step 14, but user creation must be real. Prefer completing form→DB→redirect path.

## References

- Flow: [registration.md](../../examples/flows/registration.md)
- Auth: [authentication.md](../../specs/authentication.md)
- DB: [database.md](../../architecture/database.md)

## Prerequisites

- Password service + migrations tooling.

## Scope

### In

- Migration: users table
- Repository create/get-by-email
- GET/POST /register
- Validation rules from flow doc
- Normalize email lowercase
- Terms checkbox required

### Out

- Full Mailpit verification email (step 14)
- Rate limiting (step 15)

## Implementation checklist

- [ ] users migration
- [ ] auth repository + service + handler
- [ ] register templates
- [ ] 422/error rendering for invalid input
- [ ] handler or service tests for validation

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `db/migrations/*_users.sql` | create | users |
| `internal/auth/*` | modify | registration |
| `web/templates/pages/register.html` | create | form |

## Technical notes

Unique email constraint at DB. Display name 2–50. Password ≥12. Do not reveal timing differences excessively — keep simple for v1.

## Acceptance criteria

- [ ] Valid registration inserts a user with hashed password
- [ ] Duplicate email rejected
- [ ] Invalid payloads show field errors (HTML)
- [ ] Email stored normalized lowercase
- [ ] GET /register renders form

## Verification

```bash
go run ./cmd/migrate up
go test ./internal/auth/...
# manual: register via form
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(auth): add user registration with validation
```

**Body:**

```text
Persist users with normalized emails and hashed passwords so accounts
can be created through the public registration flow.

STEP-11
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-12

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

If session not yet created on register, note that step 12 must wire it. User columns: ____.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-12.
