# Step 18: Membership & role authorization

| Field | Value |
| ----- | ----- |
| ID | `STEP-18` |
| Milestone | M3 — Workspaces |
| Status | `done` |
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

- [x] members migration
- [x] authz helpers
- [x] tests: viewer cannot mutate; outsider forbidden
- [x] wire middleware on workspace routes

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `db/migrations/*_workspace_members.sql` | create |  |
| `internal/member/*` | create | module |
| `internal/platform/middleware/*.go` | modify | workspace authz |

## Technical notes

Fail closed. Prefer 404 over 403 for cross-workspace enumeration if you choose — document the choice.

## Acceptance criteria

- [x] Three roles exist: Owner, Member, Viewer
- [x] Non-member cannot access workspace routes
- [x] Viewer cannot perform mutations
- [x] Owner can access settings routes (even if UI stub)
- [x] Authorization tests pass

## Verification

```bash
go test ./internal/member/...
go test ./internal/workspace/...
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(authz): add workspace membership and role checks
```

**Body:**

```text
Enforce Owner/Member/Viewer permissions and fail closed across
workspaces so authorization is real, not cosmetic.

STEP-18
```

**Required actions:**

- [x] Update `docs/implementation/STATUS.md` → `done`
- [x] Stage this step’s files + `STATUS.md`
- [x] Commit with the subject and body above
- [x] `git push -u origin HEAD`
- [x] Confirm clean / not ahead of `origin`
- [x] Stop — do not start STEP-19

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Role constants: `member.RoleOwner` / `RoleMember` / `RoleViewer` (DB: `owner`|`member`|`viewer`). 403 vs 404 policy: non-member or unknown workspace → **404** (no enumeration); member lacking capability (Viewer mutate, non-Owner settings) → **403**. Creator gets Owner membership in the same create transaction. Helpers: `member.RequireMembership`, `RequireOwner`, `RequireCanMutate`.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-19.
