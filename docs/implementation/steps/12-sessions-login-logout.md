# Step 12: Sessions, login, logout

| Field | Value |
| ----- | ----- |
| ID | `STEP-12` |
| Milestone | M2 — Authentication |
| Status | `todo` |
| Depends on | STEP-11 |
| Unlocks | STEP-13 |
| Estimated scope | L |

---

## Goal

Users can log in and out using hashed session tokens stored in PostgreSQL and an HttpOnly cookie.

## Description

Implement server-side sessions per auth spec. Login errors must be generic. Wire registration to also establish a session if not already done.

## References

- Flow: [login.md](../../examples/flows/login.md)
- Auth: [authentication.md](../../specs/authentication.md)

## Prerequisites

- Users table exists.

## Scope

### In

- Migration: sessions
- Token generation + sha256 storage
- Cookie forgeboard_session in dev
- GET/POST /login, POST /logout
- last_seen_at / expires_at / user_agent / ip fields
- Generic invalid credentials message

### Out

- requireAuth middleware (step 13)
- CSRF (step 13)

## Implementation checklist

- [ ] sessions migration
- [ ] session service
- [ ] login/logout handlers + templates
- [ ] tests for token hash behavior

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `db/migrations/*_sessions.sql` | create | sessions |
| `internal/auth/session.go` | create | sessions |
| `web/templates/pages/login.html` | create | login form |

## Technical notes

Never store raw token. Secure cookie in production; simple name in local HTTP.

## Acceptance criteria

- [ ] Successful login sets cookie and creates sessions row with token_hash only
- [ ] Logout revokes/deletes session and clears cookie
- [ ] Bad login returns 'Invalid email or password.'
- [ ] Expired/revoked sessions are rejected on load

## Verification

```bash
go test ./internal/auth/...
# manual login/logout via browser
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(auth): add session-based login and logout
```

**Body:**

```text
Server-side sessions let users stay signed in without JWTs, matching
the auth spec and unblocking protected application routes.

STEP-12
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-13

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Cookie name in dev: forgeboard_session. Session TTL: 7 days.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-13.
