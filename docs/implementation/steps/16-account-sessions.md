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

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(auth): add account session management page
```

**Body:**

```text
Let users inspect and revoke active sessions, demonstrating secure
session hygiene in the portfolio demo.

STEP-16
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-17

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

M2 complete. Next: workspaces.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-17.
