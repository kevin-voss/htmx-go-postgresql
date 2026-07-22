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

## Commit

**Subject (required):**

```text
feat(step-11): add users table and registration flow
```

**Body (optional):**

```text
Complete STEP-11 so the next agent can continue from a green tree.
```

## Handoff to next agent

If session not yet created on register, note that step 12 must wire it. User columns: ____.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-12.
