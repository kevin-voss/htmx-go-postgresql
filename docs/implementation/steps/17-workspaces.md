# Step 17: Workspaces

| Field | Value |
| ----- | ----- |
| ID | `STEP-17` |
| Milestone | M3 — Workspaces |
| Status | `todo` |
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

- [ ] workspaces migration
- [ ] CRUD-create + show
- [ ] templates
- [ ] tests for slug uniqueness

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `db/migrations/*_workspaces.sql` | create |  |
| `internal/workspace/*` | create | module |
| `web/templates/pages/workspace_*.html` | create |  |

## Technical notes

Prefer creating Owner membership in the same transaction when membership table arrives — coordinate with step 18 if split.

## Acceptance criteria

- [ ] User can create a workspace with name + slug
- [ ] Duplicate slug rejected
- [ ] GET /w/{slug} renders for existing workspace (auth required)
- [ ] Unknown slug → 404

## Verification

```bash
go test ./internal/workspace/...
make migrate
```

## Commit

**Subject (required):**

```text
feat(step-17): add workspace creation and slug routing
```

**Body (optional):**

```text
Complete STEP-17 so the next agent can continue from a green tree.
```

## Handoff to next agent

Slug validation regex: ____. Creator role wiring: ____.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-18.
