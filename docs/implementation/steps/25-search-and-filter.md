# Step 25: Issue search & filtering

| Field | Value |
| ----- | ----- |
| ID | `STEP-25` |
| Milestone | M4 — Projects & issues |
| Status | `done` |
| Depends on | STEP-24 |
| Unlocks | STEP-26 |
| Estimated scope | M |

---

## Goal

Project issue list supports search and filters (status, priority, assignee, label) via query params.

## Description

Server-rendered filtered lists. HTMX can enhance later; full page query-param filter is enough for this step.

## References

- Product: [product.md](../../specs/product.md)
- Pages: [pages-and-routes.md](../../specs/pages-and-routes.md)

## Prerequisites

- Issues + labels.

## Scope

### In

- Query param parsing
- Repository filter methods
- Filter form UI
- Empty-state when no matches

### Out

- Full-text search engine
- Saved filters

## Implementation checklist

- [x] Implement filters in repo
- [x] Wire UI
- [x] Tests for filter combinations

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `internal/issue/*` | modify | filtering |
| `web/templates/pages/issue_list.html` | modify | filter form |

## Technical notes

Keep SQL readable; watch injection — use parameterized queries only.

## Acceptance criteria

- [x] Filter by status works
- [x] Filter by assignee works
- [x] Text search matches title (and description if scoped)
- [x] Filters combine with AND semantics
- [x] No SQL injection via params

## Verification

```bash
go test ./internal/issue/...
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(issues): add search and filtering on issue lists
```

**Body:**

```text
Help users find work via query params for status, assignee, labels,
and text without a separate search product.

STEP-25
```

**Required actions:**

- [x] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-26

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

M4 complete. Supported query params: `q`, `status`, `priority`, `assignee` (`none` = unassigned), `label`.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-26.
