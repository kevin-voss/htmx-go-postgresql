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

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(css): add layered tokens and light/dark foundation
```

**Body:**

```text
Ship modern CSS layers and design tokens without a build step so UI
work can stay consistent across auth and app screens.

STEP-06
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-07

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Class naming convention: button / form-field / issue-card. Extend components.css as features land.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-07.
