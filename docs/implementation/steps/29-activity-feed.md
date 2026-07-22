# Step 29: Activity events & feed

| Field | Value |
| ----- | ----- |
| ID | `STEP-29` |
| Milestone | M6 — Portfolio quality |
| Status | `todo` |
| Depends on | STEP-28 |
| Unlocks | STEP-30 |
| Estimated scope | M |

---

## Goal

Mutating actions write activity_events in the same transaction; project/workspace activity feed is visible.

## Description

Activity history demonstrates transactional service-layer design. At least one flow (e.g. issue create or comment) must write activity in the same DB transaction.

## References

- Overview: [overview.md](../../architecture/overview.md)
- DB: [database.md](../../architecture/database.md)
- DoD: [DEFINITION_OF_DONE.md](../../DEFINITION_OF_DONE.md)

## Prerequisites

- Issues/comments exist.

## Scope

### In

- Migration: activity_events
- activity service
- Record events on create issue, status change, comment (minimum set)
- Feed UI on project or workspace page
- Prove transactional write in test

### Out

- Realtime websocket feed

## Implementation checklist

- [ ] migration
- [ ] instrument key mutations
- [ ] feed template
- [ ] transaction test

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `db/migrations/*_activity.sql` | create |  |
| `internal/activity/*` | create |  |
| `web/templates/components/activity_*.html` | create |  |

## Technical notes

Service layer owns transactions. Repositories stay free of HTTP.

## Acceptance criteria

- [ ] activity_events table populated on key actions
- [ ] Feed visible in UI
- [ ] At least one transaction writes domain row + activity together (test proves rollback behavior or joint commit)
- [ ] Events are workspace/project scoped correctly

## Verification

```bash
go test ./internal/activity/...
go test ./internal/issue/...
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(activity): add activity events and project feed
```

**Body:**

```text
Record key mutations transactionally and surface a feed so reviewers
can see collaboration history.

STEP-29
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-30

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Event types: ____. Feed location: ____.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-30.
