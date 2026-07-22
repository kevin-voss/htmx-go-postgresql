# Step 18: Membership & role authorization

| Field | Value |
| ----- | ----- |
| ID | `STEP-18` |
| Milestone | M3 — Workspaces |
| Status | `todo` |
| Depends on | STEP-17 |
| Unlocks | STEP-19 |
| Estimated scope | L |

---

## Goal

workspace_members enforces Owner/Member/Viewer; middleware blocks cross-workspace access.

## Description

Implement RBAC from specs/roles.md. All /w/{slug}/... routes must resolve membership before handler logic.

## References

- Roles: [roles.md](../../specs/roles.md)
- Middleware: [middleware.md](../../architecture/middleware.md)

## Prerequisites

- Workspaces exist.

## Scope

### In

- Migration: workspace_members
- Role enum/check constraints
- requireMembership / requireRole middleware
- Creator is Owner
- Authorization unit tests

### Out

- Invite email flow (step 20)

## Implementation checklist

- [ ] members migration
- [ ] authz helpers
- [ ] tests: viewer cannot mutate; outsider forbidden
- [ ] wire middleware on workspace routes

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `db/migrations/*_workspace_members.sql` | create |  |
| `internal/member/*` | create | module |
| `internal/platform/middleware/*.go` | modify | workspace authz |

## Technical notes

Fail closed. Prefer 404 over 403 for cross-workspace enumeration if you choose — document the choice.

## Acceptance criteria

- [ ] Three roles exist: Owner, Member, Viewer
- [ ] Non-member cannot access workspace routes
- [ ] Viewer cannot perform mutations
- [ ] Owner can access settings routes (even if UI stub)
- [ ] Authorization tests pass

## Verification

```bash
go test ./internal/member/...
go test ./internal/workspace/...
```

## Commit

**Subject (required):**

```text
feat(step-18): add workspace membership and role authorization
```

**Body (optional):**

```text
Complete STEP-18 so the next agent can continue from a green tree.
```

## Handoff to next agent

Role constants: ____. 403 vs 404 policy: ____.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-19.
