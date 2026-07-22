# Step 17: Workspaces

| Field | Value |
| ----- | ----- |
| ID | `STEP-17` |
| Milestone | M3 — Workspaces |
| Status | `done` |
| Depends on | STEP-16 |
| Unlocks | STEP-18 |
| Estimated scope | M |

---

## Goal

Authenticated users can create a workspace with a unique slug and open GET /w/{workspaceSlug}.

## Description

Introduce workspaces table and creation UI at /app/workspaces/new. Slug validation and uniqueness are required.

## References

- Product: [product.md](../../specs/product.md)
- Pages: [pages-and-routes.md](../../specs/pages-and-routes.md)
- DB: [database.md](../../architecture/database.md)

## Prerequisites

- Auth works.

## Scope

### In

- Migration: workspaces
- Create workspace service/repo/handler
- Slug rules (lowercase, unique)
- Workspace home page stub
- Creator membership may wait until step 18 — if so, create owner membership in same transaction here preferred

### Out

- Invitations
- Full settings

## Implementation checklist

- [x] workspaces migration
- [x] CRUD-create + show
- [x] templates
- [x] tests for slug uniqueness

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `db/migrations/*_workspaces.sql` | create |  |
| `internal/workspace/*` | create | module |
| `web/templates/pages/workspace_*.html` | create |  |

## Technical notes

Prefer creating Owner membership in the same transaction when membership table arrives — coordinate with step 18 if split.

## Acceptance criteria

- [x] User can create a workspace with name + slug
- [x] Duplicate slug rejected
- [x] GET /w/{slug} renders for existing workspace (auth required)
- [x] Unknown slug → 404

## Verification

```bash
go test ./internal/workspace/...
make migrate
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(workspaces): add workspace creation and slug routes
```

**Body:**

```text
Introduce workspace tenancy with unique slugs so projects and members
have a stable URL namespace.

STEP-17
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-18

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Slug validation regex: `^[a-z0-9]+(-[a-z0-9]+)*$` (length 2–48, lowercased). Creator role wiring: deferred to STEP-18 (`workspace_members` + Owner); `workspaces.created_by` already stores the creating user.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-18.
