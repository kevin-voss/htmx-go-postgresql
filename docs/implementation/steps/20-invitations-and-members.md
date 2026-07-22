# Step 20: Invitations & member management

| Field | Value |
| ----- | ----- |
| ID | `STEP-20` |
| Milestone | M3 — Workspaces |
| Status | `todo` |
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

- [ ] invitations migration
- [ ] handlers + templates
- [ ] accept for existing + new users
- [ ] authz tests for invite permissions

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

- [ ] Owner can send invitation email (visible in Mailpit)
- [ ] Accepting invite creates membership
- [ ] Members page lists members and roles
- [ ] Viewer cannot invite
- [ ] Invalid token handled cleanly

## Verification

```bash
make dev
# invite flow via Mailpit
```

## Commit

**Subject (required):**

```text
feat(step-20): add workspace invitations and member management
```

**Body (optional):**

```text
Complete STEP-20 so the next agent can continue from a green tree.
```

## Handoff to next agent

M3 complete. Invite token TTL: ____.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-21.
