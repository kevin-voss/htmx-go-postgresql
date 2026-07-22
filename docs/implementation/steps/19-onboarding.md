# Step 19: First-time onboarding

| Field | Value |
| ----- | ----- |
| ID | `STEP-19` |
| Milestone | M3 — Workspaces |
| Status | `todo` |
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

- [ ] Onboarding routes/UI
- [ ] Transactional create
- [ ] Skip onboarding if already member

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `internal/workspace or app onboarding` | create/modify |  |
| `web/templates/pages/onboarding.html` | create |  |
| `db/migrations/*_projects.sql` | create | if project created here |

## Technical notes

If projects migration is introduced here, step 21 extends rather than recreates.

## Acceptance criteria

- [ ] New user without memberships lands in onboarding
- [ ] Completing onboarding creates workspace, Owner membership, and first project
- [ ] User with memberships does not see onboarding gate
- [ ] Only required fields are asked

## Verification

```bash
# manual: fresh register → onboarding → project
```

## Commit

**Subject (required):**

```text
feat(step-19): add first-time workspace and project onboarding
```

**Body (optional):**

```text
Complete STEP-19 so the next agent can continue from a green tree.
```

## Handoff to next agent

Project table status: created in this step / deferred. Onboarding route: ____.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-20.
