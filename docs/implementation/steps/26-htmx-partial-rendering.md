# Step 26: Vendor HTMX 4 & partial rendering

| Field | Value |
| ----- | ----- |
| ID | `STEP-26` |
| Milestone | M5 — HTMX experience |
| Status | `todo` |
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

- [ ] Vendor file committed
- [ ] Helper + one dual-mode handler
- [ ] Document HX-Request-Type handling

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `web/static/vendor/htmx-4.0.0-beta5.min.js` | create | pinned |
| `internal/platform/request/*.go` | create/modify | helpers |
| `web/templates/layouts/base.html` | modify | script tag |

## Technical notes

No CDN dependency for runtime. Do not upgrade past pin without human approval.

## Acceptance criteria

- [ ] Vendored HTMX file exists at pinned version
- [ ] Layout loads local HTMX
- [ ] Partial request returns fragment without full layout chrome
- [ ] Full request returns full page

## Verification

```bash
ls web/static/vendor/htmx-4.0.0-beta5.min.js
go test ./internal/platform/...
```

## Commit

**Subject (required):**

```text
feat(step-26): vendor HTMX 4 and add partial rendering helpers
```

**Body (optional):**

```text
Complete STEP-26 so the next agent can continue from a green tree.
```

## Handoff to next agent

Pinned path: /static/vendor/htmx-4.0.0-beta5.min.js. First dual-mode route: ____.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-27.
