# Step 22: Issues core (create, list, detail)

| Field | Value |
| ----- | ----- |
| ID | `STEP-22` |
| Milestone | M4 — Projects & issues |
| Status | `todo` |
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

- [ ] issues migration
- [ ] issue service with number allocation
- [ ] pages + forms
- [ ] repository tests for numbering

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

- [ ] Creating issues increments issue_number per project
- [ ] Detail route works with issue number
- [ ] List shows issues for the project only
- [ ] Validation errors shown for empty title
- [ ] Viewer can read; cannot create

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

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-23

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Issue number allocation strategy: ____. Default status: Backlog.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-23.
