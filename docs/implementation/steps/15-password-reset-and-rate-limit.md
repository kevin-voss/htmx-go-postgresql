# Step 15: Password reset & login rate limit

| Field | Value |
| ----- | ----- |
| ID | `STEP-15` |
| Milestone | M2 — Authentication |
| Status | `todo` |
| Depends on | STEP-14 |
| Unlocks | STEP-16 |
| Estimated scope | L |

---

## Goal

Complete forgot/reset password flow exists; login is rate limited with 429 when exceeded.

## Description

Deliver the portfolio-required password-reset flow and basic login throttling. Reset emails go through Mailpit.

## References

- Auth: [authentication.md](../../specs/authentication.md)
- HTTP: [http-conventions.md](../../specs/http-conventions.md)
- DoD: [DEFINITION_OF_DONE.md](../../DEFINITION_OF_DONE.md)

## Prerequisites

- Mailer works.

## Scope

### In

- Migration: password_reset_tokens
- GET/POST /forgot-password
- GET/POST /reset-password/{token}
- Generic response on forgot (do not reveal account existence)
- Login rate limiting → 429

### Out

- Distributed rate limit store — in-memory/DB simple is fine for v1

## Implementation checklist

- [ ] Reset token lifecycle
- [ ] Forms + handlers
- [ ] Rate limiter middleware/service
- [ ] Tests for token consume-once

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `db/migrations/*_password_reset.sql` | create | tokens |
| `internal/auth/*` | modify | reset + limiter |
| `web/templates/pages/forgot_password.html` | create |  |
| `web/templates/pages/reset_password.html` | create |  |

## Technical notes

Always show the same forgot-password acknowledgment. Invalidate token after use.

## Acceptance criteria

- [ ] User can reset password via email link end-to-end
- [ ] Used/expired tokens fail safely
- [ ] Forgot-password does not disclose account existence
- [ ] Excessive login attempts return 429

## Verification

```bash
go test ./internal/auth/...
# manual Mailpit reset flow
```

## Commit

**Subject (required):**

```text
feat(step-15): add password reset and login rate limiting
```

**Body (optional):**

```text
Complete STEP-15 so the next agent can continue from a green tree.
```

## Handoff to next agent

Rate limit parameters: ____. Reset TTL: ____.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-16.
