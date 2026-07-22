# Step 26: Vendor HTMX 4 & partial rendering

| Field | Value |
| ----- | ----- |
| ID | `STEP-26` |
| Milestone | M5 — HTMX experience |
| Status | `done` |
| Depends on | STEP-25 |
| Unlocks | STEP-27 |
| Estimated scope | M |

---

## Goal

HTMX 4.0.0-beta5 is vendored; helpers distinguish partial vs full requests; at least one page supports fragment responses.

## Description

Pin and vendor HTMX. Implement isPartialRequest and dual templates for one route (e.g. project show).

## References

- HTMX decision: [htmx-decision.md](../../specs/htmx-decision.md)
- Rendering: [rendering.md](../../architecture/rendering.md)
- JS policy: [javascript-policy.md](../../specs/javascript-policy.md)

## Prerequisites

- Templates exist.

## Scope

### In

- Download/commit htmx-4.0.0-beta5.min.js under web/static/vendor/
- Include script in layout
- isPartialRequest helper
- Dual render path on one handler
- Tiny app.js only if needed (stay under budget)

### Out

- All features converted — do incrementally in following steps

## Implementation checklist

- [x] Vendor file committed
- [x] Helper + one dual-mode handler
- [x] Document HX-Request-Type handling

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `web/static/vendor/htmx-4.0.0-beta5.min.js` | create | pinned |
| `internal/platform/request/*.go` | create/modify | helpers |
| `web/templates/layouts/base.html` | modify | script tag |

## Technical notes

No CDN dependency for runtime. Do not upgrade past pin without human approval.

## Acceptance criteria

- [x] Vendored HTMX file exists at pinned version
- [x] Layout loads local HTMX
- [x] Partial request returns fragment without full layout chrome
- [x] Full request returns full page

## Verification

```bash
ls web/static/vendor/htmx-4.0.0-beta5.min.js
go test ./internal/platform/...
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(htmx): vendor HTMX 4 and partial render helpers
```

**Body:**

```text
Pin HTMX locally and distinguish full vs partial responses so later
steps can swap fragments safely.

STEP-26
```

**Required actions:**

- [x] Update `docs/implementation/STATUS.md` → `done`
- [x] Stage this step’s files + `STATUS.md`
- [x] Commit with the subject and body above
- [x] `git push -u origin HEAD`
- [x] Confirm clean / not ahead of `origin`
- [x] Stop — do not start STEP-27

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Pinned path: /static/vendor/htmx-4.0.0-beta5.min.js. First dual-mode route: `GET /w/{workspaceSlug}/projects/{projectSlug}` (`project_show` / fragment `project_content`).

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-27.
