# Step 22: Issues core (create, list, detail)

| Field | Value |
| ----- | ----- |
| ID | `STEP-22` |
| Milestone | M4 — Projects & issues |
| Status | `done` |
| Depends on | STEP-21 |
| Unlocks | STEP-23 |
| Estimated scope | L |

---

## Goal

Users can create issues with sequential issue_number per project, list them, and open detail pages.

## Description

Core issue entity with UUID id + per-project issue_number. Full-page flows first; HTMX enhancements come in M5.

## References

- Flow: [issue-creation.md](../../examples/flows/issue-creation.md)
- DB: [database.md](../../architecture/database.md)
- HTTP: [http-conventions.md](../../specs/http-conventions.md)

## Prerequisites

- Projects exist.

## Scope

### In

- Migration: issues
- Unique (project_id, issue_number)
- Allocate next issue_number safely (transaction)
- Create/list/detail handlers + templates
- Title/description validation
- Default status Backlog

### Out

- HTMX no-reload create (step 27)
- Labels
- Comments

## Implementation checklist

- [x] issues migration
- [x] issue service with number allocation
- [x] pages + forms
- [x] repository tests for numbering

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `db/migrations/*_issues.sql` | create |  |
| `internal/issue/*` | create | module |
| `web/templates/pages/issue_*.html` | create |  |
| `web/templates/components/issue_card.html` | create | fragment-ready |

## Technical notes

Display like FORGE-1 can use project key + number; store integer issue_number.

## Acceptance criteria

- [x] Creating issues increments issue_number per project
- [x] Detail route works with issue number
- [x] List shows issues for the project only
- [x] Validation errors shown for empty title
- [x] Viewer can read; cannot create

## Verification

```bash
go test ./internal/issue/...
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(issues): add issue create, list, and detail
```

**Body:**

```text
Allocate per-project issue numbers and ship core issue pages so work
can be tracked inside each project.

STEP-22
```

**Required actions:**

- [x] Update `docs/implementation/STATUS.md` → `done`
- [x] Stage this step’s files + `STATUS.md`
- [x] Commit with the subject and body above
- [x] `git push -u origin HEAD`
- [x] Confirm clean / not ahead of `origin`
- [x] Stop — do not start STEP-23

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Issue number allocation strategy: transaction locks `projects` row (`SELECT … FOR UPDATE`), then `MAX(issue_number)+1` for that `project_id`, insert with `UNIQUE (project_id, issue_number)`. Default status: `backlog` (label Backlog). Routes: `GET/POST /w/{ws}/projects/{ps}/issues`, detail `GET /w/{ws}/projects/{ps}/issues/{n}` and `GET /w/{ws}/issues/{n}` (workspace lookup 404s if number is ambiguous across projects). Viewer blocked from create via `RequireCanMutate`. Priority/assignee/archive deferred to STEP-23.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-23.
