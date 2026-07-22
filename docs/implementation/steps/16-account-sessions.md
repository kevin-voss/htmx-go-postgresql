# Step 16: Account sessions management

| Field | Value |
| ----- | ----- |
| ID | `STEP-16` |
| Milestone | M2 — Authentication |
| Status | `todo` |
| Depends on | STEP-15 |
| Unlocks | STEP-17 |
| Estimated scope | S |

---

## Goal

Authenticated users can list active sessions and revoke them at GET /account/sessions.

## Description

Close the auth milestone with session visibility — important for demonstrating secure session design.

## References

- Pages: [pages-and-routes.md](../../specs/pages-and-routes.md)
- Auth: [authentication.md](../../specs/authentication.md)

## Prerequisites

- Sessions table populated on login.

## Scope

### In

- GET /account/sessions
- List non-revoked sessions for current user (ua, ip, last_seen, current flag)
- Revoke action (POST/DELETE)
- Cannot use revoked session afterwards

### Out

- Account settings beyond sessions (prefs later)

## Implementation checklist

- [ ] Handler + template
- [ ] Revoke endpoint
- [ ] Mark current session

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `internal/auth/*` | modify | list/revoke |
| `web/templates/pages/account_sessions.html` | create |  |

## Technical notes

Do not show raw tokens. Show metadata only.

## Acceptance criteria

- [ ] Page lists sessions for the logged-in user only
- [ ] Revoking a session prevents further authenticated requests with that cookie
- [ ] Current session is identifiable
- [ ] Other users' sessions never appear

## Verification

```bash
go test ./internal/auth/...
# manual revoke check
```

## Commit

**Subject (required):**

```text
feat(step-16): add account sessions management page
```

**Body (optional):**

```text
Complete STEP-16 so the next agent can continue from a green tree.
```

## Handoff to next agent

M2 complete. Next: workspaces.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-17.
