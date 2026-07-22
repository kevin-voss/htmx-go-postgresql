# Step 25: Issue search & filtering

| Field | Value |
| ----- | ----- |
| ID | `STEP-25` |
| Milestone | M4 — Projects & issues |
| Status | `todo` |
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

- [ ] Implement filters in repo
- [ ] Wire UI
- [ ] Tests for filter combinations

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `internal/issue/*` | modify | filtering |
| `web/templates/pages/issue_list.html` | modify | filter form |

## Technical notes

Keep SQL readable; watch injection — use parameterized queries only.

## Acceptance criteria

- [ ] Filter by status works
- [ ] Filter by assignee works
- [ ] Text search matches title (and description if scoped)
- [ ] Filters combine with AND semantics
- [ ] No SQL injection via params

## Verification

```bash
go test ./internal/issue/...
```

## Commit

**Subject (required):**

```text
feat(step-25): add issue search and filtering
```

**Body (optional):**

```text
Complete STEP-25 so the next agent can continue from a green tree.
```

## Handoff to next agent

M4 complete. Supported query params: ____.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-26.
