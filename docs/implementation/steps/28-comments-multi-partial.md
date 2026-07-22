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

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(comments): add multi-target hx-partial comment posts
```

**Body:**

```text
Append comments, update counts, and clear the form in one response to
satisfy the multi-target HTMX definition-of-done requirement.

STEP-28
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-29

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

M5 complete. Comment delete policy: ____.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-29.
