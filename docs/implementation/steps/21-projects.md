# Step 21: Projects

| Field | Value |
| ----- | ----- |
| ID | `STEP-21` |
| Milestone | M4 — Projects & issues |
| Status | `todo` |
| Depends on | STEP-20 |
| Unlocks | STEP-22 |
| Estimated scope | M |

---

## Goal

Members can create and view projects under a workspace via slug routes.

## Description

Solidify projects module: list, create, show. Archive project as Owner capability if straightforward.

## References

- Pages: [pages-and-routes.md](../../specs/pages-and-routes.md)
- DB: [database.md](../../architecture/database.md)
- Structure: [project-structure.md](../../architecture/project-structure.md)

## Prerequisites

- Workspaces + roles.

## Scope

### In

- projects table (if not from onboarding)
- GET/POST project routes
- project slug unique per workspace
- project list + detail templates
- RBAC: Viewer read; Member+ create

### Out

- Issues
- Board drag-drop

## Implementation checklist

- [ ] project module complete
- [ ] routes under /w/{ws}/projects...
- [ ] tests for slug uniqueness per workspace

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `internal/project/*` | create/modify |  |
| `web/templates/pages/project_*.html` | create |  |
| `db/migrations/*_projects.sql` | create/modify |  |

## Technical notes

Path: /w/{workspaceSlug}/projects/{projectSlug}

## Acceptance criteria

- [ ] Create project within workspace
- [ ] Open project page by slug
- [ ] Non-member cannot access
- [ ] Duplicate project slug in same workspace rejected

## Verification

```bash
go test ./internal/project/...
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(projects): add projects within workspaces
```

**Body:**

```text
Let members create and open projects under a workspace slug as the
container for issues and boards.

STEP-21
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-22

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Project model fields: ____.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-22.
