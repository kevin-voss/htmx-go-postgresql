# Step 19: First-time onboarding

| Field | Value |
| ----- | ----- |
| ID | `STEP-19` |
| Milestone | M3 — Workspaces |
| Status | `done` |
| Depends on | STEP-18 |
| Unlocks | STEP-20 |
| Estimated scope | M |

---

## Goal

Users without memberships are guided to create a workspace + first project with minimal fields.

## Description

Implement the onboarding flow. Creating the first project may introduce a thin projects table early or stub until step 21 — prefer real project row if small.

## References

- Flow: [onboarding.md](../../examples/flows/onboarding.md)
- Roles: [roles.md](../../specs/roles.md)

## Prerequisites

- Memberships work.

## Scope

### In

- Detect no membership after login/register
- Onboarding form: workspace name, slug, first project name
- Transaction: workspace + owner membership + project
- Redirect to project page

### Out

- Rich workspace switcher polish

## Implementation checklist

- [x] Onboarding routes/UI
- [x] Transactional create
- [x] Skip onboarding if already member

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `internal/workspace or app onboarding` | create/modify |  |
| `web/templates/pages/onboarding.html` | create |  |
| `db/migrations/*_projects.sql` | create | if project created here |

## Technical notes

If projects migration is introduced here, step 21 extends rather than recreates.

## Acceptance criteria

- [x] New user without memberships lands in onboarding
- [x] Completing onboarding creates workspace, Owner membership, and first project
- [x] User with memberships does not see onboarding gate
- [x] Only required fields are asked

## Verification

```bash
# manual: fresh register → onboarding → project
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(onboarding): add first-time workspace and project setup
```

**Body:**

```text
Guide users without memberships through a minimal create flow so the
happy path reaches a usable project quickly.

STEP-19
```

**Required actions:**

- [x] Update `docs/implementation/STATUS.md` → `done`
- [x] Stage this step’s files + `STATUS.md`
- [x] Commit with the subject and body above
- [x] `git push -u origin HEAD`
- [x] Confirm clean / not ahead of `origin`
- [x] Stop — do not start STEP-20

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Project table status: **created** in this step (`db/migrations/00008_projects.sql`). Onboarding route: `GET/POST /app/onboarding`. Login/register redirect to `/app`, which gates users without memberships to onboarding. Completing onboarding creates workspace + Owner membership + project in one transaction and redirects to `/w/{workspaceSlug}/projects/{projectSlug}` (thin `project_show` page). Project slug is derived from the project name via `project.SlugFromName`. STEP-21 should extend the project module (list/create/archive) rather than recreating the table.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-20.
