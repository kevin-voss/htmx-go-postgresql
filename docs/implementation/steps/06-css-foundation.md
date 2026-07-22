# Step 06: CSS foundation

| Field | Value |
| ----- | ----- |
| ID | `STEP-06` |
| Milestone | M1 — Foundation |
| Status | `todo` |
| Depends on | STEP-05 |
| Unlocks | STEP-07 |
| Estimated scope | M |

---

## Goal

Modern layered CSS with tokens, light/dark via light-dark()/color-scheme, and base component stubs exists and is linked from the layout.

## Description

Implement the CSS strategy from specs/css.md without Tailwind or a build step. Focus on tokens and structure; page-specific polish can grow later.

## References

- CSS spec: [css.md](../../specs/css.md)
- JS policy: [javascript-policy.md](../../specs/javascript-policy.md)

## Prerequisites

- Layout can link stylesheets.

## Scope

### In

- reset, tokens, base, layout, components, utilities CSS files
- @layer ordering
- oklch + light-dark tokens
- Link stylesheets in base layout
- prefers-reduced-motion basic respect

### Out

- Full page visual design for every feature
- View transitions everywhere

## Implementation checklist

- [ ] Add CSS files under web/static/css
- [ ] Wire into base layout
- [ ] Verify no build tooling introduced

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `web/static/css/*.css` | create | layered CSS |
| `web/templates/layouts/base.html` | modify | link tags |

## Technical notes

No Tailwind/Sass/PostCSS/npm. Keep selectors readable.

## Acceptance criteria

- [ ] CSS loads on a rendered page
- [ ] Tokens define background/text/primary/spacing/radius
- [ ] No CSS build step in Makefile/Docker
- [ ] color-scheme supports light and dark

## Verification

```bash
grep -R "@layer" web/static/css
go build ./...
```

## Commit

**Subject (required):**

```text
feat(step-06): add modern CSS layers and design tokens
```

**Body (optional):**

```text
Complete STEP-06 so the next agent can continue from a green tree.
```

## Handoff to next agent

Class naming convention: button / form-field / issue-card. Extend components.css as features land.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-07.
