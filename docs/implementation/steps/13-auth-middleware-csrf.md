# Step 13: Auth middleware & CSRF

| Field | Value |
| ----- | ----- |
| ID | `STEP-13` |
| Milestone | M2 — Authentication |
| Status | `done` |
| Depends on | STEP-12 |
| Unlocks | STEP-14 |
| Estimated scope | M |

---

## Goal

Session loading, requireAuthentication, and CSRF protection are wired; /app is protected.

## Description

Build the middleware chain foundations from architecture/middleware.md. Public routes remain accessible; authenticated area requires a valid session. CSRF must cover state-changing POSTs.

## References

- Middleware: [middleware.md](../../architecture/middleware.md)
- Pages: [pages-and-routes.md](../../specs/pages-and-routes.md)

## Prerequisites

- Sessions work.

## Scope

### In

- loadSession middleware
- requireAuthentication
- CSRF token issuance + validation
- GET /app protected dashboard stub
- Security headers middleware (basic)

### Out

- Workspace authz (step 18)
- Full panic/request-id polish if not done — add if quick

## Implementation checklist

- [ ] Implement middleware packages
- [ ] Protect /app
- [ ] Add CSRF hidden fields to forms
- [ ] Tests: unauthenticated /app redirects or 401

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `internal/platform/middleware/*.go` | create | chain |
| `internal/auth/middleware.go` | create | requireAuth |
| `internal/app/routes.go` | modify | chains |

## Technical notes

Decide redirect-to-login vs 401 for browser vs HTMX; document choice. Prefer redirect for full pages.

## Acceptance criteria

- [ ] Unauthenticated GET /app redirects to /login (or documented equivalent)
- [ ] Authenticated user can open /app
- [ ] POST without CSRF token is rejected
- [ ] Session user available in request context

## Verification

```bash
go test ./internal/platform/middleware/...
go test ./internal/auth/...
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(auth): add auth middleware and CSRF protection
```

**Body:**

```text
Protect /app and state-changing forms so unauthenticated and forged
requests cannot reach privileged handlers.

STEP-13
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-14

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Context helpers: `auth.UserFromContext` / `auth.SessionFromContext`. CSRF field name: `csrf_token` (cookie `forgeboard_csrf` in dev). Unauthenticated full-page requests redirect to `/login` (302); HTMX requests get 401.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-14.
