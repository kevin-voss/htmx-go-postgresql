# Step 20: Invitations & member management

| Field | Value |
| ----- | ----- |
| ID | `STEP-20` |
| Milestone | M3 — Workspaces |
| Status | `done` |
| Depends on | STEP-19 |
| Unlocks | STEP-21 |
| Estimated scope | L |

---

## Goal

Owners can invite users by email; invitees accept via /invites/{token}; members page lists memberships.

## Description

Implement invitation flow end-to-end with Mailpit. Include basic member list and role change/remove as Owner capabilities allow.

## References

- Flow: [invitation.md](../../examples/flows/invitation.md)
- Roles: [roles.md](../../specs/roles.md)
- Pages: [pages-and-routes.md](../../specs/pages-and-routes.md)

## Prerequisites

- Mailer + memberships.

## Scope

### In

- Migration: workspace_invitations
- Invite create + email
- GET /invites/{token} accept path
- GET /w/{slug}/members
- Owner remove member / change role (Viewer/Member)
- Workspace settings stub page OK

### Out

- Admin role
- Bulk invite CSV

## Implementation checklist

- [x] invitations migration
- [x] handlers + templates
- [x] accept for existing + new users
- [x] authz tests for invite permissions

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `db/migrations/*_invitations.sql` | create |  |
| `internal/member/*` | modify | invites |
| `web/templates/pages/members.html` | create |  |
| `web/templates/pages/invite_accept.html` | create |  |

## Technical notes

Default invited role: Member. Owner cannot be removed by non-owners.

## Acceptance criteria

- [x] Owner can send invitation email (visible in Mailpit)
- [x] Accepting invite creates membership
- [x] Members page lists members and roles
- [x] Viewer cannot invite
- [x] Invalid token handled cleanly

## Verification

```bash
make dev
# invite flow via Mailpit
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(workspaces): add invitations and member management
```

**Body:**

```text
Allow owners to invite collaborators by email and manage memberships
through the Mailpit-backed invite accept path.

STEP-20
```

**Required actions:**

- [x] Update `docs/implementation/STATUS.md` → `done`
- [x] Stage this step’s files + `STATUS.md`
- [x] Commit with the subject and body above
- [x] `git push -u origin HEAD`
- [x] Confirm clean / not ahead of `origin`
- [x] Stop — do not start STEP-21

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

M3 complete. Invite token TTL: **7 days** (`member.InvitationTTL()`). Default invited role: Member. Accept path: `GET /invites/{token}` auto-accepts when signed in with matching email; login/register support `?next=` for return. Owners manage roles (Member/Viewer) and remove non-owners on `/w/{slug}/members`.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-21.
