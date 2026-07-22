# Step 28: Comments with multi-target hx-partial

| Field | Value |
| ----- | ----- |
| ID | `STEP-28` |
| Milestone | M5 — HTMX experience |
| Status | `todo` |
| Depends on | STEP-27 |
| Unlocks | STEP-29 |
| Estimated scope | L |

---

## Goal

Users can comment on issues; response uses multiple <hx-partial> targets (list, count, cleared form).

## Description

Implement comments module and the canonical multi-target update from the comments flow — required by definition of done.

## References

- Flow: [comments.md](../../examples/flows/comments.md)
- DoD: [DEFINITION_OF_DONE.md](../../DEFINITION_OF_DONE.md)
- Rendering: [rendering.md](../../architecture/rendering.md)

## Prerequisites

- Issue detail page + HTMX helpers.

## Scope

### In

- Migration: issue_comments
- POST create comment
- DELETE comment (author or elevated role)
- Response with multiple hx-partial blocks
- Viewer cannot comment

### Out

- Rich text
- Reactions

## Implementation checklist

- [ ] comments migration + module
- [ ] templates for comment + form + count
- [ ] multi-partial response
- [ ] tests for create authz

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `db/migrations/*_comments.sql` | create |  |
| `internal/comment/*` | create |  |
| `web/templates/fragments/comment_*.html` | create |  |

## Technical notes

Match the example structure in comments.md closely.

## Acceptance criteria

- [ ] Comment posts via HTMX and appends without full reload
- [ ] Comment count updates in same response
- [ ] Form clears in same response
- [ ] Uses <hx-partial> multi-target pattern
- [ ] Viewer cannot create comments

## Verification

```bash
go test ./internal/comment/...
# manual multi-target update
```

## Commit

**Subject (required):**

```text
feat(step-28): add comments with multi-target hx-partial updates
```

**Body (optional):**

```text
Complete STEP-28 so the next agent can continue from a green tree.
```

## Handoff to next agent

M5 complete. Comment delete policy: ____.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-29.
