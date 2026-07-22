# Step 27: 422 fragments & inline HTMX updates

| Field | Value |
| ----- | ----- |
| ID | `STEP-27` |
| Milestone | M5 — HTMX experience |
| Status | `todo` |
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

- [ ] Update issue forms with HTMX attrs
- [ ] Return fragments from handlers
- [ ] 422 fragment for validation
- [ ] Manual UX verification

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `web/templates/**` | modify | hx-* attrs |
| `internal/issue/handler.go` | modify | fragment responses |
| `web/static/css/**` | modify | indicator styles |

## Technical notes

Error responses must be HTML fragments. Prefer explicit inheritance attributes if needed in HTMX 4.

## Acceptance criteria

- [ ] Create issue via HTMX without full page reload
- [ ] Invalid create returns 422 HTML into error target
- [ ] Status change updates UI without full reload
- [ ] Submit control disables while request in flight
- [ ] Loading indicator visible during request

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

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-28

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Fragment template names: ____. hx-status mapping documented.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-28.
