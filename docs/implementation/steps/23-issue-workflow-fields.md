# Step 23: Issue status, priority, assignee, archive

| Field | Value |
| ----- | ----- |
| ID | `STEP-23` |
| Milestone | M4 — Projects & issues |
| Status | `todo` |
| Depends on | STEP-22 |
| Unlocks | STEP-24 |
| Estimated scope | M |

---

## Goal

Issues support status workflow, priority, assignee, and archive via PATCH/POST actions (buttons/select — no drag-drop).

## Description

Implement workflow fields from issue-status flow. Prefer select/buttons. Full page refresh OK until HTMX steps.

## References

- Flow: [issue-status.md](../../examples/flows/issue-status.md)
- Roles: [roles.md](../../specs/roles.md)
- HTTP: [http-conventions.md](../../specs/http-conventions.md)

## Prerequisites

- Issues core done.

## Scope

### In

- Statuses: Backlog, Todo, In Progress, Done
- Priorities defined (e.g. low/medium/high/urgent)
- Assignee to workspace member
- Archive issue
- PATCH routes as in conventions

### Out

- Drag and drop board
- HTMX swap polish (step 27)

## Implementation checklist

- [ ] Update schema if needed
- [ ] Handlers for status/priority/assignee/archive
- [ ] UI controls on detail + list
- [ ] authz tests

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `internal/issue/*` | modify |  |
| `web/templates/**` | modify | controls |
| `db/migrations/*_issue_fields.sql` | create | if columns missing |

## Technical notes

No drag-and-drop JS.

## Acceptance criteria

- [ ] Status can change among the four values
- [ ] Priority can be set
- [ ] Assignee can be set to a workspace member or cleared
- [ ] Archived issues hidden from default lists
- [ ] Viewer cannot change fields

## Verification

```bash
go test ./internal/issue/...
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(issues): add status, priority, assignee, and archive
```

**Body:**

```text
Support the v1 workflow with buttons/selects so issues can move and
be owned without drag-and-drop complexity.

STEP-23
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-24

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Priority enum: ____. Archive semantics: soft flag.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-24.
