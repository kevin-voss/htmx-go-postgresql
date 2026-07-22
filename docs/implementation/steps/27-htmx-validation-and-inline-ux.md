# Step 27: 422 fragments & inline HTMX updates

| Field | Value |
| ----- | ----- |
| ID | `STEP-27` |
| Milestone | M5 — HTMX experience |
| Status | `done` |
| Depends on | STEP-26 |
| Unlocks | STEP-28 |
| Estimated scope | L |

---

## Goal

Issue create/update flows work via HTMX without full reload; validation returns 422 HTML fragments; loading/disabled submit behaviors exist.

## Description

Make the issue UX feel dynamic: hx-post/hx-patch, hx-status:422, swap issue cards, status changes without reload. Include loading indicators and disabled submit controls.

## References

- HTMX: [htmx-decision.md](../../specs/htmx-decision.md)
- Flow: [issue-creation.md](../../examples/flows/issue-creation.md)
- HTTP: [http-conventions.md](../../specs/http-conventions.md)
- Flow status: [issue-status.md](../../examples/flows/issue-status.md)

## Prerequisites

- HTMX vendored.
- Issues CRUD exist.

## Scope

### In

- hx-status:422 error targets
- Inline create appends card
- Status change swaps card/fragment
- hx-indicator / disable-on-request patterns
- Browser history where appropriate (hx-push-url on key navigations)

### Out

- Comments multi-partial (step 28)

## Implementation checklist

- [x] Update issue forms with HTMX attrs
- [x] Return fragments from handlers
- [x] 422 fragment for validation
- [x] Manual UX verification

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `web/templates/**` | modify | hx-* attrs |
| `internal/issue/handler.go` | modify | fragment responses |
| `web/static/css/**` | modify | indicator styles |

## Technical notes

Error responses must be HTML fragments. Prefer explicit inheritance attributes if needed in HTMX 4.

## Acceptance criteria

- [x] Create issue via HTMX without full page reload
- [x] Invalid create returns 422 HTML into error target
- [x] Status change updates UI without full reload
- [x] Submit control disables while request in flight
- [x] Loading indicator visible during request

## Verification

```bash
# manual browser HTMX flows
go test ./internal/issue/...
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(htmx): add 422 fragments and inline issue updates
```

**Body:**

```text
Validate and update issues without full reloads, using HTMX status
swaps and loading/disabled submit affordances.

STEP-27
```

**Required actions:**

- [x] Update `docs/implementation/STATUS.md` → `done`
- [x] Stage this step’s files + `STATUS.md`
- [x] Commit with the subject and body above
- [x] `git push -u origin HEAD`
- [x] Confirm clean / not ahead of `origin`
- [x] Stop — do not start STEP-28

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Fragment template names: `issue_form_errors`, `issue_list_item`, `issue_list_results`, `issue_status_panel` (plus existing `issue_card`). Create form uses `hx-status:422="target:#issue-form-errors swap:innerHTML"`; success appends `issue_list_item` into `#issue-list`. Status forms use `hx-patch` + card/`issue_status_panel` outerHTML swap. Filter form uses `hx-push-url="true"` into `#issue-list-results`. Loading via `hx-indicator` + `.htmx-indicator` CSS; submit disable via `hx-disable`.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-28.
