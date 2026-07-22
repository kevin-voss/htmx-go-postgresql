# Step 13: Auth middleware & CSRF

| Field | Value |
| ----- | ----- |
| ID | `STEP-13` |
| Milestone | M2 — Authentication |
| Status | `todo` |
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

## Commit

**Subject (required):**

```text
feat(step-13): add auth middleware and CSRF protection
```

**Body (optional):**

```text
Complete STEP-13 so the next agent can continue from a green tree.
```

## Handoff to next agent

Context key for user: ____. CSRF field name: ____.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-14.
