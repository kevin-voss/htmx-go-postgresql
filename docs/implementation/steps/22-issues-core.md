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

## Commit

**Subject (required):**

```text
feat(step-22): add issue creation, list, and detail
```

**Body (optional):**

```text
Complete STEP-22 so the next agent can continue from a green tree.
```

## Handoff to next agent

Issue number allocation strategy: ____. Default status: Backlog.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-23.
