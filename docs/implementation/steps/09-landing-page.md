# Step 09: Landing page

| Field | Value |
| ----- | ----- |
| ID | `STEP-09` |
| Milestone | M1 — Foundation |
| Status | `todo` |
| Depends on | STEP-08 |
| Unlocks | STEP-10 |
| Estimated scope | S |

---

## Goal

GET / renders a public Forgeboard landing page with clear CTAs to register/login (pages may 404 until auth steps).

## Description

Ship the first real HTML page. Brand the product name Forgeboard prominently. Links to /register and /login can exist even if those routes arrive next milestone.

## References

- Product: [product.md](../../specs/product.md)
- Pages: [pages-and-routes.md](../../specs/pages-and-routes.md)
- CSS: [css.md](../../specs/css.md)

## Prerequisites

- Templates + CSS + make dev work.

## Scope

### In

- GET / handler
- Landing template using base layout
- Links: Register, Login
- Responsive basic layout

### Out

- Auth forms
- Marketing extras / stats strips (avoid clutter)

## Implementation checklist

- [ ] Implement home handler
- [ ] Create landing template + page CSS if needed
- [ ] Verify via browser or curl HTML

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `internal/app or platform handlers` | modify | home |
| `web/templates/pages/home.html` | create | landing |
| `web/static/css/pages/*.css` | create/modify | optional |

## Technical notes

Keep the first viewport simple: brand, one headline, short support line, CTA group — match product positioning, not a dashboard.

## Acceptance criteria

- [ ] GET / returns 200 HTML containing the word Forgeboard
- [ ] CTAs to /register and /login are present
- [ ] Page uses base layout and CSS tokens
- [ ] Works when opened via make dev

## Verification

```bash
curl -s localhost:8080/ | grep -i forgeboard
```

## Commit

**Subject (required):**

```text
feat(step-09): add public landing page
```

**Body (optional):**

```text
Complete STEP-09 so the next agent can continue from a green tree.
```

## Handoff to next agent

M1 complete. Next: password service and registration.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-10.
